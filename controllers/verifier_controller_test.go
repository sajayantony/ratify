package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1alpha1 "github.com/deislabs/ratify/api/v1alpha1"
)

var _ = Describe("Verifier controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		VerifierObjectName      = "verifier-notary"
		VerifierObjectNamespace = "default"
		NotaryVerifierName      = "notaryv2"
		NotaryArtifactType      = "application/vnd.cncf.notary.v2.signature"
		timeout                 = time.Second * 10
		duration                = time.Second * 10
		interval                = time.Millisecond * 250
	)

	Context("When adding new verifier", func() {
		It("New verifiers should have been created successfully", func() {
			By("By creating a new verifier with empty parameters")
			ctx := context.Background()
			verifier := &configv1alpha1.Verifier{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "config.ratify.deislabs.io/v1alpha1",
					Kind:       "Verifier",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      VerifierObjectName,
					Namespace: VerifierObjectNamespace,
				},
				Spec: configv1alpha1.VerifierSpec{
					Name:          NotaryVerifierName,
					ArtifactTypes: NotaryArtifactType,
					/*parameters: {  todo: test rawString properties
						verificationCerts: []string{"/usr/local/ratify-certs"},
					},*/
				},
			}
			Expect(k8sClient.Create(ctx, verifier)).Should(Succeed())

			/*
				we want to test the internal object created not the spec properties
					verifierLookupKey := types.NamespacedName{Name: VerifierObjectName, Namespace: VerifierObjectNamespace}
					createdVerifier := &configv1alpha1.Verifier{}

					// We'll need to retry given that creation may not immediately happen.
					Eventually(func() bool {
						err := k8sClient.Get(ctx, verifierLookupKey, createdVerifier)
						if err != nil {
							return false
						}
						return true
					}, timeout, interval).Should(BeTrue())

					Expect(createdVerifier.Spec.Name).Should(Equal("1 * * * *"))*/
		})
	})

})
