# OpenVPN Access server
Provides a web frontend with OpenID Connect authentication that can create and sign new openvpn client certificates. The client certificates and ca.crt/ca.key are stored in S3. An ovpn config is generated and offered as a download. The client crt/key can be encrypted (at rest) using AWS KMS.

# Configuration
OAUTH2\_CLIENT\_ID=
OAUTH2\_CLIENT\_SECRET=
OAUTH2\_REDIRECT\_URL=http://url/callback
OAUTH2\_URL=https://url/oidc
CSRF\_KEY=32-byte-long-auth-key
CLIENT\_CERT\_ORG=org
S3\_BUCKET=
S3\_PREFIX=
S3\_KMS\_ARN=
AWS\_REGION=
