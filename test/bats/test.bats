#!/usr/bin/env bats

load helpers

BATS_TESTS_DIR=${BATS_TESTS_DIR:-test/bats/tests}
WAIT_TIME=60
SLEEP_TIME=1

@test "quick start test" {
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5
    run kubectl run demo --image=ratify.azurecr.io/testimage:signed
    assert_success
    run kubectl run demo1 --image=ratify.azurecr.io/testimage:unsigned
    assert_failure

    echo "cleaning up"
    kubectl delete pod demo
    run kubectl delete ns demo
}

@test "verifier CRD CRUD test" {
    run kubectl apply -f ./config/samples/config_v1alpha1_verifier_licensechecker.yaml
    assert_success
    
    run kubectl run demo2 --image=ratify.azurecr.io/testimage:signed
    echo "Deployment should fail as license checker does not verify signature"
    assert_failure

    run kubectl apply -f ./config/samples/config_v1alpha1_verifier_notary.yaml
    assert_success

    run kubectl run demo2 --image=ratify.azurecr.io/testimage:signed
    assert_success

    run kubectl delete -f ./config/samples/config_v1alpha1_verifier_notary.yaml
    assert_success

    run kubectl run demo2 --image=ratify.azurecr.io/testimage:signed
    assert_failure

    echo "cleaning up"
    run kubectl delete -f ./config/samples/config_v1alpha1_verifier_licensechecker.yaml
    run kubectl delete -f ./config/samples/config_v1alpha1_verifier_notary.yaml
    run kubectl delete ns demo2
    run kubectl delete pod demo2
}

@test "configmap update test" {
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5
    run kubectl run demo2 --image=ratify.azurecr.io/testimage:signed
    assert_success

    run kubectl get configmaps ratify-configuration --namespace=ratify-service -o yaml > currentConfig.yaml
    run kubectl delete -f ./library/default/samples/constraint.yaml
                                            
    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl replace --namespace=ratify-service -f ${BATS_TESTS_DIR}/configmap/invalidconfigmap.yaml"
    echo "Waiting for 150 second for configuration update"
    sleep 150

    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    run kubectl run demo3 --image=ratify.azurecr.io/testimage:signed
    echo "Current time after validate : $(date +"%T")"
    assert_failure
     
    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl replace --namespace=ratify-service -f currentConfig.yaml"
}