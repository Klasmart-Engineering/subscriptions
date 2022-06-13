# Subscriptions Service

### Local Setup

##### First Time?

```
  brew install tilt
  brew install k3d
  brew install kubectx 
  brew install istioctl
  brew install jq
  brew install helm
  go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.11.0
  
  make setup-k3d
```

 - Go to https://github.com/settings/tokens.
 - Click "Generate new token"
 - Name it "Container Registry", select read:packages from the list, click "Generate Token"
 - Copy the key from the next page

```
echo "github-username:the-token" | base64
```

Copy the output

```
echo '{"auths":{"ghcr.io":{"auth":"paste-output-here"}}}' | base64
```

Create a file named imagepull.yaml in your home directory:

```
kind: Secret
type: kubernetes.io/dockerconfigjson
apiVersion: v1
metadata:
  name: dockerconfigjson-github-com
  labels:
    app: app-name
data:
  .dockerconfigjson: output-of-last-command-here
```

Now add the secret to your k3d cluster, start it if it's not already running (see below) then run

```
kubectl apply -f imagepull.yaml
```

Go back to https://github.com/settings/tokens, click on the "Configure SSO" dropdown and click "Authorize" next to KL-Engineering.

##### To run locally in K3d:

```
  make run-k3d
```

##### To run locally in docker

```
  make run-docker
```

##### To remote debug locally in docker

```
  make run-docker-debug
```
- Then add the following configuration in Goland (TODO also add instructions for VScode)
![img.png](readme-images/img.png)

Or to debug the instance in K3D, connect to port 40002 instead.

##### To run unit tests

```
  make test-unit
```

##### To run integration tests

```
  make test-integration
```

### Profiles

Add `-profile=profile-name` to the command line or `PROFILE=profile-name` as an environment variable to select a profile when running.  The config is then loaded from the relevant json file in the profiles directory.

Values can be overriden by environment variables by using an underscore to traverse the JSON structure, e.g. `SERVER_PORT=1234` will override the Server.Port config value.

### Open API

Endpoint boilerplate is generated from openapi-spec.yaml.

```
make openapi-generate
```

This generates src/api/api.gen.go.  This contains an interface which you need to implement in api_impl.go
