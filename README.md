# OpenVPN Access server
Provides a web frontend with OpenID Connect authentication that can create and sign new openvpn client certificates. The client certificates and ca.crt/ca.key are stored in S3. An ovpn config is generated and offered as a download. The client crt/key can be encrypted (at rest) using AWS KMS.

# Configuration

| Environment Variable | Description |
| -------------------- | ----------- |
| OAUTH2\_CLIENT\_ID | client id |
| OAUTH2\_CLIENT\_SECRET | client secret |
| OAUTH2\_REDIRECT\_URL | callback, e.g. http://url/callback |
| OAUTH2\_URL | oidc url, e.g. https://url/oidc |
| CSRF\_KEY | 32-byte-long-auth-key |
| CLIENT\_CERT\_ORG | organisation |
| STORAGE_TYPE | s3 or azblob (azure blob storage), default is s3 |
| S3\_BUCKET | s3 bucket where openvpn config is stored |
| S3\_PREFIX | s3 prefix, e.g. openvpn |
| S3\_KMS\_ARN | KMS ARN to encrypt s3 objects |
| AWS\_REGION | AWS Region |
| AZ_STORAGE_ACCOUNT_NAME | azure storage account name (when storage type azure) |
| AZ_STORAGE_ACCOUNT_KEY | azure storage account key. Leave empty for Managed Service Identity (MSI) (when storage type azure) |
| AZ_STORAGE_ACCOUNT_CONTAINER | azure storage account container (when storage type azure) |
