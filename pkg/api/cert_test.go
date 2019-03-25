package api

import (
	"fmt"
	"testing"
)

const caKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA5taFx8RTckpGa0W+tik+CHzOOAFEC/MqnPlGCGmle5IkISC3
JQ/3QRjcN6ryWtfBT6DTTR5qcaJxodAI9THViIzkBDwFtzi1/VhmY+G+/Eq5KmW1
MtrivNM8rtdWOsAZ9qkM8g8TDsaR8HXhYheSjcPs3bqbXLGD2DPfUMNg9CXqFrO4
7mtfF+H8pii0a60S4w4KfqdGkT+XhA8sfdaNDK5CYv3sv6136SB27TYxubn3MwW6
0cphQl+NJG47EtWw9YzTZ60j2IXlVQgyuXVSbygmO5EtoMg7SccS3NMHOqm1WBoh
BI22o96eY40iLCmB+OTPabMthNnk++uvJLBj0QIDAQABAoIBAQDHcMtvOLncQj4r
SrwmiOWW0bYef0F6jaLgsyKF/DlE1ZQvpyN1eyDmdoM1+ZWhVU4o7UwDEmdnPLGu
254RsvfKHeiXnABYOEBM3obAf3fSZQEsl2mBwcoq2NtSOKzA397O1Wpg7RNLLddD
iaWsaa9umrvZKvQ2lwzRzKeCEPyAk2kRkJdKovaEm06FBZ2L4mkMMJaAoLTnxXYI
hg2Q32HdjFdM5YfyAyad5Tqt4LneVqcTjWNkUdU1sXpqQo9+NV1MEtIxkZS10yOR
1fK3rdNVtu66eBbNVCPEVk03Cyn8Y9UPXb8c7KznGplHzV6FmRoPP3rHoZlQHvfQ
O/MPKMINAoGBAP0CSHBD5yEVN0m10hfXNg5BLHMZC68VRlTzYzkTvYfQiIm18Xao
W+wimQI9BoV9IAFgHtRJU8nzvV+z2J9FkRDsa3sXkaAQK4Zi0GpSOvJ047V19/TE
3PLtRuRUMvBZgGLm4HJpgYHR8z0oj9lV4wCJPVLggnii4zNIhPKvhjEzAoGBAOmR
I/orTxKyBjAhyQsdZF+N2YgXhuMWx7IcTCEsU2TaSDIZUkzHcFKxBdY/zXmYYe2R
PCWN7svmK28cQsqo7YhFR2JykY0bno6KTCAbFvhJnKsbMFpuW9hGm+zqvcCb/Sqo
8vp7E0BDlBM1G+Zwc8OAeZVTI2/zQ8ZzrVLAzd7rAoGBALtPYF+09b1ZXqg0csjx
rHRbLdQ8W5kQcBReaDwOcEfHS/5f89b8B6nHZ23vzg8vtm0uQ0S40M53o+DhXeN0
dlSII35q0YYl0oNYTqIYJMnxXc+u+ZZ91HII1m4eI+Qq7tDJyqsJjzaUP7cse2rU
mg2AjST5T76OIRSLgNnGtttfAoGBAOH2K7M8MPyqRDhOhzx8i/2xsvDpqfKKuFmE
7NXvFyLr1oq5WpizHeSyJC55fWUU2jDGoETIwmx62ixdT/TWZy69r2j74/p67POD
slAhRSChvrL+09G5EJv0+6bCFx9/Cfc6ig9wAFjcyCWo7LwMsMJDydyAGTmWREx6
3wS/SKxPAoGAPv5XMPi9AYEvlVjrgTrVjCj5SdBuWQBdEP9qWwwLSDx2CT1Cytdd
KlH3ioA9+Zfhu6ClD2XgzP1DNQbk2IX2cwVnjghbfGOwPifTGWVhw3BvF/iG+V7G
twgmKVLdN/fFCaOi1TWjtkdKg4FT6hRItIHua8PiMXgcsbER3MgsGqI=
-----END RSA PRIVATE KEY-----`

const caCert = `-----BEGIN CERTIFICATE-----
MIIDSzCCAjOgAwIBAgIUDOYmkb5229YWiiVNayuP/0ZX8z0wDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAwwLRWFzeS1SU0EgQ0EwHhcNMTkwMzI1MDk0MTQ4WhcNMjkw
MzIyMDk0MTQ4WjAWMRQwEgYDVQQDDAtFYXN5LVJTQSBDQTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAObWhcfEU3JKRmtFvrYpPgh8zjgBRAvzKpz5Rghp
pXuSJCEgtyUP90EY3Deq8lrXwU+g000eanGicaHQCPUx1YiM5AQ8Bbc4tf1YZmPh
vvxKuSpltTLa4rzTPK7XVjrAGfapDPIPEw7GkfB14WIXko3D7N26m1yxg9gz31DD
YPQl6hazuO5rXxfh/KYotGutEuMOCn6nRpE/l4QPLH3WjQyuQmL97L+td+kgdu02
Mbm59zMFutHKYUJfjSRuOxLVsPWM02etI9iF5VUIMrl1Um8oJjuRLaDIO0nHEtzT
BzqptVgaIQSNtqPenmONIiwpgfjkz2mzLYTZ5PvrrySwY9ECAwEAAaOBkDCBjTAd
BgNVHQ4EFgQUYqAor1cxN+7RXLAKCvNtnGwijyswUQYDVR0jBEowSIAUYqAor1cx
N+7RXLAKCvNtnGwijyuhGqQYMBYxFDASBgNVBAMMC0Vhc3ktUlNBIENBghQM5iaR
vnbb1haKJU1rK4//RlfzPTAMBgNVHRMEBTADAQH/MAsGA1UdDwQEAwIBBjANBgkq
hkiG9w0BAQsFAAOCAQEA4lMQIaNFqOG1gc1suob733J7u6ybjzPbppvngLVVWh4w
a9SMJjRadvyCKHfpSazziRohV/HUgQyaowkfE4aMAtfOzbsc0cDD3392mCBwm5Vn
vKyVOUbrOo+DfUS5IJgUHJYfG6hw1Q8WykA8v+jxHgX56i8fB7sLhgfUJ4vmdfJj
kMnPHbeQZsfeIt2qnMpAm1VbAWRy5PzBfkBU6IBW6KX0R+Jt84eoamHC8JnMuBHa
SzTeTXpi4znT6zHiBIpaer4Zz/tNChpO/h50h0I7uuwAnIdP8A8pUrDXGZoWSZzV
eSlOnpeX4wLHUHuJqM+QkBl3DUIjOcJkevPrPPzlyw==
-----END CERTIFICATE-----`

func TestCreateCert(t *testing.T) {
	fmt.Printf("starting test")
	var err error
	c := NewCert()
	parsedCaCert, err := c.readCert(caCert)
	if err != nil {
		t.Errorf("Parsed CA Error: %s", err)
		return
	}
	parsedCaKey, err := c.readPrivateKey(caKey)
	if err != nil {
		t.Errorf("Parsed CA Key Error: %s", err)
		return
	}
	clientCert, clientKey, err := c.createClientCert(parsedCaCert, parsedCaKey, "test-subject")
	if err != nil {
		t.Errorf("Create Cert error: %s", err)
		return
	}

	fmt.Printf("New Client Cert:\n%s\n\nNew Client key:\n%s", clientCert.String(), clientKey.String())
}
