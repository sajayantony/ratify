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

		})
	})

})
