load('ext://git_resource', 'git_checkout')
load('ext://helm_resource', 'helm_resource', 'helm_repo')

if not os.path.exists('../microgateway-base-helm'):
    git_checkout('git@github.com:KL-Engineering/microgateway-base-helm.git', '../microgateway-base-helm')

docker_build('local-go-image', '.',
    dockerfile='Dockerfile-debug',
    )

helm_resource('krakend', '../microgateway-base-helm/charts', flags = ['--set', 'global.imagePullSecrets[0].name=dockerconfigjson-github-com'])

docker_build('local-postgres-image', '.',
    dockerfile='Dockerfile-postgres')
k8s_yaml('./environment/local/go.yaml')
k8s_yaml('./environment/local/postgres.yaml')

k8s_yaml('./environment/local/cronjob.yaml')
k8s_resource('go-app', labels=['subscriptions'], port_forwards=['8000:8080', '40002:40000'], resource_deps=['postgres'])
k8s_resource('postgres', labels=['subscriptions'], port_forwards=5432)
