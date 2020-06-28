# ots-secret-generator
OTS secret generator is designed to create random passwords and automatically stores them to OTS service (https://github.com/onetimesecret/onetimesecret) using its API. 

After generation, it prints plain text passwords and links to its corresponding OTS URL. It helps to generate secrets when you need to share passwords with a large number of different users.

The generator was tested on local OTS instance version 0.10.1.-49f3761b-en

## Quickstart
0. Have a running instance of OTS
1. Create Account and generate an API key on OTS service
2. Download the binary depending on your host system. You can download pre compiled binaries from this repo
    - ots-secret-generator : Linux executable
    - ots-secret-generator.exe : Windows executable
3. Setup configuration file
4. Run the binary with flags that point to the configuration file and number of passwords you wish to generate. 

Example of successful run (on Linux):

```bash
./ots-secret-generator --config ./config.json --passwords 3 
```

```bash
Configuration successfully loaded.
OTS service reachable. Healthcheck response: {"status":"nominal","locale":"en"}
Starting secret generation ...
s0Xs0X -> https://ots.domain.tld/secret/m14kkk0gdy4e879gxfjc4lt2rhn4rz
oW0oW0 -> https://ots.domain.tld/secret/r0uakuru8j5qgrn2w2hff1sadb0japi
$%8$%8 -> https://ots.domain.tld/secret/s1oue7ryzrtnivoo0o5oty515hnj2h8
```

Generator acepts --config flag and --passwords flag that defines number of passwords/secrets to generate.

## Configuration
Generator checks for configuration file (default: config.json) that is provided with --config argument. Below you can see a valid configuration file. 

NOTE: All fields are required.

```json
{
    "endpoint" : "https://ots.domain.tld",
    "username" : "user@domain.tld",
    "api-key" : "df30e824932493f1ace2aa2493249369eea9008",
    "secret-ttl": 36000,
    "password-length" : 3
}
```

- **endpoint** : location where OTS is located. Only write protocol (HTTP/HTTPS) and hostname. The path to API is automatically appended
- **username**: OTS username
- **api-key**: API key that was generated for your user 
- **secret-ttl**: The remaining time (in seconds) that the secret has left to live inside OTS service before is beeing read
- **password-length**: Length of the generated password

## Additional info
- Password is generated from following character sets:
    - abcdedfghijklmnopqrst
	- ABCDEFGHIJKLMNOPQRSTUVWXYZ
    - !@#$%&*
	- 0123456789

- Maximum number of passwords generated at once is **150**

- Default values for CLI flags
    - --config config.json
    - --passwords 1
