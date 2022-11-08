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
    run kubectl run demo --image=wabbitnetworks.azurecr.io/test/net-monitor:signed
    assert_success
    run kubectl run demo1 --image=wabbitnetworks.azurecr.io/test/net-monitor:unsigned
    assert_failure

    echo "cleaning up"
    run kubectl delete pod demo
}

@test "validate crd add, replace and delete" {    
    skip "TODO: setup local test registry to enable better CRD e2e testing, default crd should installed the notary verifier"
    echo "adding license checker, delete notary verifier and validate deployment fails due to missing notary verifier"
    run kubectl apply -f ./config/samples/config_v1alpha1_verifier_licensechecker.yaml
    run kubectl delete verifiers.config.ratify.deislabs.io/verifier-notary
    assert_success
    run kubectl run demo --image=ratify.azurecr.io/testimage:signed
    assert_failure

    echo "Add notary verifier and validate deployment succeeds"
    run kubectl apply -f ./config/samples/config_v1alpha1_verifier_notary.yaml
    assert_success
    run kubectl run demo --image=ratify.azurecr.io/testimage:signed
    assert_success

    echo "cleaning up"
    run kubectl delete pod demo
}

@test "configmap update test" {
    skip "Skipping test for now as we are no longer watching for configfile update in a k8 environment.This test ensures we are watching config file updates in a non-kub scenario"
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5
    run kubectl run demo2 --image=wabbitnetworks.azurecr.io/test/net-monitor:signed
    assert_success

    run kubectl get configmaps ratify-configuration --namespace=ratify-service -o yaml > currentConfig.yaml
    run kubectl delete -f ./library/default/samples/constraint.yaml
                                            
    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl replace --namespace=ratify-service -f ${BATS_TESTS_DIR}/configmap/invalidconfigmap.yaml"
    echo "Waiting for 150 second for configuration update"
    sleep 150

    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    run kubectl run demo3 --image=wabbitnetworks.azurecr.io/test/net-monitor:signed
    echo "Current time after validate : $(date +"%T")"
    assert_failure
     
    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl replace --namespace=ratify-service -f currentConfig.yaml"
}