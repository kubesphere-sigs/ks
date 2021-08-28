pipeline {
  agent any

  parameters {
    string defaultValue: 'rick', description: 'just for testing', name: 'name', trim: true
    booleanParam defaultValue: false, description: 'You can use this flag to debug your Pipeline', name: 'debug'
    choice choices: ['v1.18.8', 'v1.18.9'], description: 'Please choose the target Kubernetes version', name: 'kubernetesVersion'
  }

  environment {
    APP_NAME = "this is a sample app"
  }

  stages{
    stage('simple'){
      steps{
        echo "My name is ${params.name}."
        echo "Target Kubernetes version is " + params.kubernetesVersion
        echo "env " + env.APP_NAME

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
}
