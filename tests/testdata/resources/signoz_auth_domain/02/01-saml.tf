# Scenario 02 — saml auth domain. Exercises the saml config variant, whose spec
# requires an X.509 signing certificate. Create, no-drift, destroy.
resource "signoz_auth_domain" "scenario_02" {
  name    = "saml.scenario-02.example.com"
  enabled = true
  config = {
    saml = {
      kind = "saml"
      spec = {
        entity_id = "https://idp.example.com/metadata"
        location  = "https://idp.example.com/sso"
        certificate = <<EOT
-----BEGIN CERTIFICATE-----
MIIDFTCCAf2gAwIBAgIUM9Jwyl0+RqvCV+x6qdVx0XSjIR0wDQYJKoZIhvcNAQEL
BQAwGjEYMBYGA1UEAwwPaWRwLmV4YW1wbGUuY29tMB4XDTI2MDgxNDIwMDM1OFoX
DTM2MDgxMTIwMDM1OFowGjEYMBYGA1UEAwwPaWRwLmV4YW1wbGUuY29tMIIBIjAN
BgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4R4YIRdyn9ZDTgHpdxl6oInv76Ot
hxCJJu0jLVjrJr78u66z1CDTOCFRS21wvNzuMtecmOP70woWtDmlOnab8fgknbgV
VBaKFcpBcncORceFr+OzERs/E//xw+LmsjvjnsPEsAiMSRhoUJUTgvKwJRvrSX1k
KwqCvYFvFcvXmG0KYH+YKsI9WsWe+SGq+2zAxsCwj58ERf6WWTC8uOzKSNG19h8J
k3tIFfzJLyrhNoP1qgBdMUa3+WDOkkdUVQTsA+0SldAOveCLi7CaRyF8mE0gXql1
/OO/QMpQsShHdU9RExKuUHc0P/lkLtJkG5ymL6XRqNKst20ze5B/ZFQccQIDAQAB
o1MwUTAdBgNVHQ4EFgQUjlyZ4/Rt+YU7VbieVDiMKWHWYfQwHwYDVR0jBBgwFoAU
jlyZ4/Rt+YU7VbieVDiMKWHWYfQwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0B
AQsFAAOCAQEAeKRb12T+UZRdz4kzfwrnhhuhkQdQoQq8aQxTb7bSXG8dAjve/HZA
OCFNwvu/cY8CLCUyFdpWFSMBRBv5PEyqo5fdyCo1ME9Sb/2WGDRugj5zXL3Hg+j3
d6mqHPNzyKHrnAol8lZJvL79AuRa8OMu0SPhTdT7W7+8DDJWhJelDdMg6wWvUVuH
vsx8L5eXbiAwUvfsgc1VJ1fNX+zcwgWwijpQ0wc50ohTdDkg7V1Kh2KHeC44FDI4
zsF2s3d7Kul9uljftByDKl4Zw3bKsxhH20p1nTShAr7Th9lfA8g+trcZmSSQxRVp
upKEsDwUxIBG/x35zEpXiv58QoECU0DLIA==
-----END CERTIFICATE-----
EOT
      }
    }
  }
}
