docker_build('local-go-image', '.',
    dockerfile='Dockerfile-local',
    )

docker_build('local-postgres-image', '.',
    dockerfile='Dockerfile-postgres')
k8s_yaml('./environment/local/go.yaml')
k8s_yaml('./environment/local/postgres.yaml')
k8s_yaml('./environment/local/krakend/krakend-config.yml')
k8s_yaml('./environment/local/krakend/krakend-deployment.yml')
k8s_yaml('./environment/local/krakend/krakend-service.yml')
k8s_yaml('./environment/local/cronjob.yaml')
k8s_resource('go-app', labels=['subscriptions'], port_forwards=8000, resource_deps=['postgres'])
k8s_resource('krakend-deployment', labels=['api-gateway'], port_forwards=8010)
k8s_resource('postgres', labels=['subscriptions'], port_forwards=5432)
