properties([disableConcurrentBuilds()])

//11
node("dind") {
    container('dind') {
        compat.test_build()
    }
}
