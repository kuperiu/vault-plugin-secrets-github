# Vault GitHub Secrets Plugin

The GitHub backend is a secret engine for GitHub apps private keys that genrates ephemeral installation tokens.

## Usage
1. Setup Vault:
    - Download the latest release from [here](https://github.com/kuperiu/vault-plugin-secrets-github/releases)
    - Extract the binary file from the archive and copy it to the plugin directory
    - Change your vault config with the follwing line: (example)

        ```plugin_directory = "/etc/vault/vault_plugins"``` 
    - Start Vault
2. Enable the secret engine with: ```vault secrets enable github```
3. Create a GitHub app:
    - Create a private key and download it
    - Create an installation for the app
4. Configure the backend:
    - Create a json file with the following scheme with the private key, app id and installation id
      ```json
        {
            "app_id": 123,
            "installation_id": 456,
            "key": "-----BEGIN RSA PRIVATE KEY..."
        }
      ```
      - Post the json file with:
        ```curl -H "Content-Type: application/json" -H "X-Vault-Token: root" -X POST  http://${VAULT_ADDR}/v1/github/org -d @payload.json -v```
5. Get the installation token with: ```vault read github/org```
6. You should see the token in the following output:
    ```
    Key                    Value
    ---                    -----
    ExpireAt               2020-10-21T10:31:55Z
    Permissions            map[actions:write administration:write checks:write contents:write deployments:write issues:write metadata:read packages:write pages:write pull_requests:write repository_hooks:write repository_projects:admin secrets:write security_events:write statuses:write vulnerability_alerts:read workflows:write]
    RepositorySelection    all
    Token                  v1.19e3d9de4f21a2349e72a54f45149625f48e8848
    ```
    

## Remarks
1. To authenticate with the token you have to use **token** instaed of **Bearer** in the Authorization header. Please see [here](https://docs.github.com/en/free-pro-team@latest/developers/apps/authenticating-with-github-apps#authenticating-as-an-installation) 
2. You can find the token available endpoints [here](https://docs.github.com/en/free-pro-team@latest/rest/overview/endpoints-available-for-github-apps)
