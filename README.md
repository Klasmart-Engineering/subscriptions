# Subscriptions POC

### Local Setup

##### First Time?

```
  brew install tilt
  brew install k3d
  brew install kubectx 
  brew install istioctl
  brew install jq
  brew install helm
  
  mkdir -p /tmp/k3dvol/  
  k3d cluster create factory --image rancher/k3s:v1.20.15-k3s1 --volume /tmp/k3dvol:/tmp/k3dvol --registry-create local-factory-registry -p "30001:30001@loadbalancer"
  kubectl create ns subscriptions
  kubens subscriptions
  kubectl label namespace subscriptions istio-injection=enabled
  
```

##### To run locally in K3d:

```
  ./run.sh
```

##### To run locally in docker

```
  ./run-docker.sh
```

##### To remote debug locally in docker

```
  ./run-docker.sh debug
```
- Then add the following configuration in Goland (TODO also add instructions for VScode)
![img.png](img.png)

##### To run unit tests

```
  go test -v ./test/unit/...
```

##### To run integration tests

```
  ./run-integration-tests.sh
```

##### Profiles

Add `-profile=profile-name` to the command line or `PROFILE=profile-name` as an environment variable to select a profile when running.  The config is then loaded from the relevant json file in the profiles directory.

Values can be overriden by environment variables by using an underscore to traverse the JSON structure, e.g. `SERVER_PORT=1234` will override the Server.Port config value.
