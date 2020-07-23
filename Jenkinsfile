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
    options {
        skipDefaultCheckout()
        buildDiscarder(logRotator(numToKeepStr: '10', artifactNumToKeepStr: '10'))
    }
    environment {
        // Get maven to download as many artifacts at a time as possible
        MAVEN_OPTS = "-Dmaven.artifact.threads=30"
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
                    sh "docker push hub.sinimini.com/docker/moneytree:latest"
                }
            }
        }
    }
}