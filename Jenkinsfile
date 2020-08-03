// Cancel previous builds
def buildNumber = env.BUILD_NUMBER as int
if (buildNumber > 1) milestone(buildNumber - 1)
milestone(buildNumber)

pipeline {
    agent {
        kubernetes {
            yamlFile "jenkins-containers.yaml"
        }
    }
    stages {
        stage('Build Container') {
            steps {
                container('docker') {
                    sh "docker build -t hub.sinimini.com/docker/moneytree:latest ."
                }
            }
        }

        stage('Push Container') {
            steps {
                container('docker') {
                    withCredentials([usernamePassword(credentialsId: "hub", usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
                        sh "docker login -u $USERNAME -p $PASSWORD hub.sinimini.com"
                    }
                    sh "docker push hub.sinimini.com/docker/moneytree:latest"
                }
            }
        }
    }
    post {
        success {
            build job: 'SinisterMinister/automation/master', wait: false
        }
    }
}