# One Pod at a time

This is a test repository.

## Deploy the ValidationWebhook on OpenShift

1. Deploy the webhook service

    ~~~sh
    oc create -f deploy/webhook-svc-deployment.yaml
    ~~~
2. Update the `CA_BUNDLE` for the webhooks

    ~~~sh
    deploy/updatecabundle.sh deploy/validatingwebhook.yaml
    ~~~
3. Deploy the `ValidatingWebhookConfiguration`:

    ~~~sh
    oc create -f deploy/validatingwebhook.yaml
    ~~~
4. Test the webhook:

    ~~~sh
    oc create -f deploy/test-validating/test-deployment.yaml
    ~~~
5. Clean everything:

    ~~~sh
    oc delete ns one-pod-at-a-time test-ns-onepodatatime
    ~~~

    ~~~sh
    oc delete -f deploy/webhook-svc-deployment.yaml 
    ~~~

    ~~~sh
    oc delete -f deploy/validatingwebhook.yaml
    ~~~
