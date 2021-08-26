pipeline {
  agent {
    node {
      label 'go'
    }
  }
  stages {
    stage('Code Clone') {
      steps {
        git(url: 'https://github.com/kubesphere-sigs/demo-go-http', changelog: true, poll: false)
      }
    }
    stage('Test & Code Scan') {
      steps {
        container('go') {
          sh 'go test ./... -coverprofile=covprofile'
          withCredentials([string(credentialsId : 'sonar-token' ,variable : 'TOKEN' ,)]) {
            withSonarQubeEnv('sonar') {
              sh 'sonar-scanner -Dsonar.login=$TOKEN'
            }
          }
        }

        waitForQualityGate 'false'
      }
    }
    stage('Build Image & Push') {
      steps {
        container('go') {
          sh '''    CGO_ENABLED=0 GOARCH=amd64 go build -o bin/go-server -ldflags "-w"
    chmod u+x bin/go-server'''
          withCredentials([usernamePassword(credentialsId : 'rick-docker-hub' ,passwordVariable : 'PASS' ,usernameVariable : 'USER' ,)]) {
            sh 'echo "$PASS" | docker login -u "$USER" --password-stdin'
            sh '''cat <<EOM >Dockerfile
FROM alpine
COPY bin/go-server go-server
EXPOSE 80
ENTRYPOINT ["go-server"]
EOM
docker build . -t surenpi/go-demo
docker push surenpi/go-demo'''
          }
        }
      }
    }
  }
}
