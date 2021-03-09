package pipeline

var jenkinsfileTemplateForSimple = `
pipeline {
  agent any
  
  parameters {
    string defaultValue: 'rick', description: 'just for testing', name: 'name', trim: true
    booleanParam defaultValue: false, description: 'You can use this flag to debug your Pipeline', name: 'debug'
    choice choices: ['v1.18.8', 'v1.18.9'], description: 'Please choose the target Kubernetes version', name: 'kubernetesVersion'
  }

  stages{
    stage('simple'){
      steps{
        echo "My name is ${params.name}."
        echo "Target Kubernetes version is " + params.kubernetesVersion

        script {
          if ("${params.debug}" == "true") {
            echo "You can put some debug info at here."
            echo "Another way to use param: " + params.name
          }
        }
      }
    }

    stage('debug stage') {
      when {
        equals expected: true, actual: params.debug
      }
      steps {
        echo "It's joke, there're debug info here."

        script {
          result = input message: 'Please input your name!', ok: 'Confirm',
			  parameters: [string(defaultValue: 'rick',
				description: 'This should not be your real name.', name: 'name', trim: true)]
          echo result
        }
      }
    }
    
    stage('parallel'){
      parallel {
        stage('channel-1'){
          steps{
            input message: 'Please input your age!', ok: 'Confirm',
              parameters: [string(defaultValue: '18',
                description: 'Just a joke.', name: 'age', trim: true)]
          }
        }
        stage('channel-2'){
          steps{
            input message: 'Please input your weight!', ok: 'Confirm',
              parameters: [string(defaultValue: '50',
                description: 'Just a joke.', name: 'weight', trim: true)]
          }
        }
      }
    }
  }
}`

var jenkinsfileTemplateForJava = `
pipeline {
  agent {
    node {
      label 'maven'
    }
  }
  stages {
    stage('Clone') {
      steps {
        git(url: 'https://github.com/kubesphere-sigs/demo-java', changelog: true, poll: false)
      }
    }
    stage('Build & Test') {
      steps {
        container('maven') {
          sh 'mvn package test'
        }
      }
    }
    stage('Code Scan') {
      steps {
        withSonarQubeEnv('sonar') {
          container('maven') {
            sh '''mvn --version
mvn sonar:sonar \\
  -Dsonar.projectKey=test \\
  -Dsonar.host.url=http://139.198.9.130:30687/ \\
  -Dsonar.login=b3e146cdb76ecef5ffb12743779cd78e69a4b5c5'''
          }

        }

        waitForQualityGate 'false'
      }
    }
    stage('Build Image') {
      steps {
        container('maven') {
          withCredentials([usernamePassword(credentialsId : 'docker' ,passwordVariable : 'PASS' ,usernameVariable : 'USER' ,)]) {
            sh '''docker login -u $USER -p $PASS
cat <<EOM >Dockerfile
FROM kubesphere/java-8-centos7:v2.1.0
COPY target/demo-java-1.0.0.jar demo.jar
COPY target/lib demo-lib
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "demo.jar"]
EOM
docker build . -t surenpi/java-demo
docker push surenpi/java-demo'''
          }
        }
      }
    }
  }
}
`

var jenkinsfileTemplateForGo = `
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
`
