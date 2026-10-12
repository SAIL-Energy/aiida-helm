# External Keycloak

AIIDA delegates authentication to Keycloak. The chart configures the application
and browser client but does not install, import, or administer a Keycloak realm.
Use an HTTPS provider reachable by browsers and by AIIDA.

## Realm and client

Create a realm and a public OpenID Connect client. Enable Standard Flow, require
PKCE `S256`, and disable Direct Access Grants for the production browser client.
Use the exact AIIDA origin for web origins and restrict redirect URIs to its
application paths.

The following realm import is a starting point. Replace both instances of
`https://aiida.example.com` before importing it. Provision users through your
identity process after import.

```json
{
  "realm": "AIIDA",
  "enabled": true,
  "sslRequired": "external",
  "registrationAllowed": false,
  "clients": [
    {
      "clientId": "AIIDA",
      "enabled": true,
      "protocol": "openid-connect",
      "publicClient": true,
      "standardFlowEnabled": true,
      "directAccessGrantsEnabled": false,
      "redirectUris": ["https://aiida.example.com/*"],
      "webOrigins": ["https://aiida.example.com"],
      "attributes": {"pkce.code.challenge.method": "S256"}
    }
  ]
}
```

For the [internal installation example](../examples/values.internal.yaml), use
`http://localhost:8080/*` for `redirectUris` and `http://localhost:8080` for
`webOrigins` instead. These values must match the URL used in the browser.
Replace them with the public HTTPS origin when enabling ingress.

Configure the chart with matching identifiers:

```yaml
keycloak:
  externalHost: https://keycloak.example.com
  internalHost: ""
  realm: AIIDA
  clientId: AIIDA
```

`externalHost` is the public base URL. Do not append `/realms/AIIDA`. The issuer
is derived from the public base URL and realm. `internalHost` can provide a
different base URL for retrieving signing keys, provided it serves the same
realm. Leaving it empty uses the public URL.

## Verification

Verify discovery from a browser:

```text
https://keycloak.example.com/realms/AIIDA/.well-known/openid-configuration
```

Ensure the discovery document's issuer matches the public realm URL, and that
AIIDA can reach the configured realm signing-key endpoint with a trusted
certificate. Complete an actual UI login after installing AIIDA.

Repeated logouts or redirect failures usually require checking the public AIIDA
URL, issuer, realm, client ID, redirect URIs, web origins, and reverse proxy
settings. The chart's public client has no client secret.

See the [official AIIDA Keycloak guide](https://architecture.eddie.energy/aiida/1-running/keycloak.html)
and [Keycloak administration guide](https://www.keycloak.org/docs/latest/server_admin/)
for provider configuration. Upstream Docker examples may use development
credentials and different client settings. Do not copy their development
passwords into a public deployment.
