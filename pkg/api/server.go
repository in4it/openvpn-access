package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/in4it/openvpn-access/pkg/storage"
)

/*
 * Config has a Port property
 */
type Config struct {
	Port string
}

type server struct {
	config       Config
	auth         *Auth
	sessionStore *sessions.CookieStore
}

type response struct {
	Message string `json:"message"`
}
type errorResponse struct {
	Message string `json:"message"`
}

/*
 * NewServer initializes new server
 */
func NewServer(conf Config) *server {
	return &server{
		config: conf,
		auth:   NewAuth(),
	}
}
func (s *server) Start() {
	r := mux.NewRouter()

	prefix := os.Getenv("URL_PREFIX")
	prefixRoot := prefix
	if prefix == "" {
		prefixRoot = "/"
	}

	if prefix != "/" {
		r.HandleFunc("/", s.rootHandler)
	}

	r.HandleFunc(prefixRoot, s.homeHandler)
	r.HandleFunc(prefix+"/login", s.loginHandler)
	r.HandleFunc(prefix+"/callback", s.callbackHandler)
	r.HandleFunc(prefix+"/ovpnconfig", s.ovpnConfigHandler)

	http.Handle("/", r)

	// initialize auth
	err := s.auth.init()
	if err != nil {
		log.Fatalf("Could not initialize auth: %s", err)
	}

	// initialize session store
	s.sessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

	// enable csrf
	CSRF := csrf.Protect([]byte(os.Getenv("CSRF_KEY")))

	// enable logging
	loggedRouter := handlers.LoggingHandler(os.Stdout, CSRF(r))

	// start server
	fmt.Printf("Starting server on port %s with prefix %s\n", s.config.Port, prefix)
	log.Fatal(http.ListenAndServe(":"+s.config.Port, loggedRouter))
}

func (s *server) rootHandler(w http.ResponseWriter, r *http.Request) {
	var response response
	response.Message = "app up and running"
	json.NewEncoder(w).Encode(response)

}

func (s *server) homeHandler(w http.ResponseWriter, r *http.Request) {
	var response response
	response.Message = "running"
	json.NewEncoder(w).Encode(response)

}

func (s *server) loginHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.auth.getAuthURL(csrf.Token(r)), http.StatusFound)
}
func (s *server) callbackHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "token-session")

	token, err := s.auth.getToken(r.URL.Query().Get("code"))
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: err.Error()})
		return
	}

	// save token
	session.Values["token"] = token
	session.Save(r, w)

	// Parse and verify ID Token payload.
	err = s.auth.verifyToken(token)
	if err != nil {
		// handle error
		json.NewEncoder(w).Encode(errorResponse{Message: err.Error()})
		return
	}

	http.Redirect(w, r, os.Getenv("URL_PREFIX")+"/ovpnconfig", 301)
}

func (s *server) ovpnConfigHandler(w http.ResponseWriter, r *http.Request) {
	var (
		clientCert bytes.Buffer
		clientKey  bytes.Buffer
	)
	year := time.Now().Format("2006")
	session, err := s.sessionStore.Get(r, "token-session")
	if err != nil || session.Values["token"] == nil {
		// handle error
		json.NewEncoder(w).Encode(errorResponse{Message: "Unauthorized"})
		return
	}

	err = s.auth.verifyToken(session.Values["token"].(string))
	if err != nil {
		// handle error
		json.NewEncoder(w).Encode(errorResponse{Message: err.Error()})
		return
	}

	login := s.auth.getLogin()
	if login == "" {
		json.NewEncoder(w).Encode(errorResponse{Message: "getLogin error"})
		return
	}

	// check in storage if .crt / .key is already created
	blobStorage, storageBucket, storagePrefix, err := s.getStorage()
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "Could not create session: " + err.Error()})
		return
	}
	err = blobStorage.HeadObject(storageBucket, storagePrefix+"issued/client-"+login+"-"+year+".crt")
	if err == nil {
		clientCert, _ = blobStorage.GetObject(storageBucket, storagePrefix+"issued/client-"+login+"-"+year+".crt")
		clientKey, _ = blobStorage.GetObject(storageBucket, storagePrefix+"private/client-"+login+"-"+year+".key")
	}

	// retrieve CA key / crt
	caKey, err := blobStorage.GetObject(storageBucket, storagePrefix+"private/ca.key")
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "ca.crt download error: " + err.Error()})
		return
	}
	caCert, err := blobStorage.GetObject(storageBucket, storagePrefix+"ca.crt")
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "ca.crt download error: " + err.Error()})
		return
	}
	taKey, err := blobStorage.GetObject(storageBucket, storagePrefix+"ta.key")
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "ta.key download error: " + err.Error()})
		return
	}
	// create new cert (if not cached)
	if clientCert.Len() == 0 || clientKey.Len() == 0 {
		c := NewCert()
		parsedCaCert, err := c.readCert(caCert.String())
		if err != nil {
			fmt.Printf("cert: %s", caCert.String())
			json.NewEncoder(w).Encode(errorResponse{Message: "Parsed CA cert Error: " + err.Error()})
			return
		}
		parsedCaKey, err := c.readPrivateKey(caKey.String())
		if err != nil {
			json.NewEncoder(w).Encode(errorResponse{Message: "Parsed CA key Error: " + err.Error()})
			return
		}
		clientCert, clientKey, err = c.createClientCert(parsedCaCert, parsedCaKey, login)
		if err != nil {
			json.NewEncoder(w).Encode(errorResponse{Message: "Create Cert error: " + err.Error()})
			return
		}
		// write key and cert to Blob Storage
		err = blobStorage.PutObject(storageBucket, storagePrefix+"issued/client-"+login+"-"+year+".crt", clientCert.String(), os.Getenv("S3_KMS_ARN"))
		if err != nil {
			json.NewEncoder(w).Encode(errorResponse{Message: "Blob Storage Put error: " + err.Error()})
			return
		}
		err = blobStorage.PutObject(storageBucket, storagePrefix+"private/client-"+login+"-"+year+".key", clientKey.String(), os.Getenv("S3_KMS_ARN"))
		if err != nil {
			json.NewEncoder(w).Encode(errorResponse{Message: "Blob Storage Put error: " + err.Error()})
			return
		}
	}

	// output openvpn config
	ovpnConfig, err := blobStorage.GetObject(storageBucket, storagePrefix+"openvpn-client.conf")
	strOvpnConfig := ovpnConfig.String()
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "openvpn-client.conf download error: " + err.Error()})
	}
	strOvpnConfig = strings.Replace(strOvpnConfig, "[CERT]", clientCert.String(), -1)
	strOvpnConfig = strings.Replace(strOvpnConfig, "[KEY]", clientKey.String(), -1)
	strOvpnConfig = strings.Replace(strOvpnConfig, "[CA]", caCert.String(), -1)
	strOvpnConfig = strings.Replace(strOvpnConfig, "[TLS-AUTH]", taKey.String(), -1)

	// write to client
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Type", "application/force-download")
	w.Header().Set("Content-Type", "application/download")
	w.Header().Set("Content-Disposition", "attachment; filename=client-"+strings.Replace(login, "@", "-", -1)+".ovpn")
	fmt.Fprintf(w, strOvpnConfig)
}
func (s *server) getStorage() (storage.StorageIf, string, string, error) {
	// azure storage
	if os.Getenv("STORAGE_TYPE") == "azblob" {
		if os.Getenv("AZ_STORAGE_ACCOUNT_KEY") != "" {
			blobStorage, err := storage.NewAzBlob(os.Getenv("AZ_STORAGE_ACCOUNT_NAME"), os.Getenv("AZ_STORAGE_ACCOUNT_KEY"))
			return blobStorage, os.Getenv("AZ_STORAGE_ACCOUNT_CONTAINER"), "", err
		}
		// with MSI
		blobStorage, err := storage.NewAzBlobWithMSI(os.Getenv("AZ_STORAGE_ACCOUNT_NAME"))
		return blobStorage, os.Getenv("AZ_STORAGE_ACCOUNT_CONTAINER"), "", err
	}
	// default storage
	blobStorage, err := storage.NewS3()
	return blobStorage, os.Getenv("S3_BUCKET"), os.Getenv("S3_PREFIX") + "/pki/", err
}
