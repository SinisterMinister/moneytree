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
        stage("Build Images") {
            parallel {
                stage('Build Moneytree') {
                    steps {
                        container('docker') {
                            dir('cmd/moneytree') {
                                sh "docker build --build-arg branch=$BRANCH_NAME -t hub.sinimini.com/docker/moneytree:latest ."
                                sh "docker tag hub.sinimini.com/docker/moneytree:latest hub.sinimini.com/docker/moneytree:$BRANCH_NAME"
                            }
                        }
                    }
                }

                stage('Build MiracleGrow') {
                    steps {
                        container('docker') {
                            dir('cmd/miraclegrow') {
                                sh "docker build --build-arg branch=$BRANCH_NAME -t hub.sinimini.com/docker/miraclegrow:latest ."
                                sh "docker tag hub.sinimini.com/docker/miraclegrow:latest hub.sinimini.com/docker/miraclegrow:$BRANCH_NAME"
                            }
                        }
                    }
                }
            }
        }
        stage("Push Images") {
            parallel {
                stage('Push Latest Images') {
                    when {
                        branch 'master'
                    }
                    steps {
                        container('docker') {
                            withCredentials([usernamePassword(credentialsId: "hub", usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
                                sh 'docker login -u $USERNAME -p $PASSWORD hub.sinimini.com'
                            }
                            sh "docker push hub.sinimini.com/docker/moneytree:latest"
                            sh "docker push hub.sinimini.com/docker/miraclegrow:latest"
                        }
                    }
                }

                stage('Push Branch Images') {
                    when {
                        not {
                            branch 'master'
                        }
                    }
                    steps {
                        container('docker') {
                            withCredentials([usernamePassword(credentialsId: "hub", usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
                                sh 'docker login -u $USERNAME -p $PASSWORD hub.sinimini.com'
                            }
                            sh "docker push hub.sinimini.com/docker/moneytree:$BRANCH_NAME"
                            sh "docker push hub.sinimini.com/docker/miraclegrow:$BRANCH_NAME"
                        }
                    }
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