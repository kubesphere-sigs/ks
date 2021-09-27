pipeline {
    agent any
    stages {
        stage('stage-1') {
            parallel {
                stage('stage-1-1') {
                    steps {
                        echo 'stage-1-1'
                    }
                }
                stage('stage-1-2') {
                    steps {
                        echo 'stage-1-2'
                    }
                }
            }
        }
        stage('stage-2') {
            steps {
                echo 'stage-2'
            }
        }
    }
}
