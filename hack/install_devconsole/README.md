##Non-admin user viewing the console with developer console perspective

Refers [https://jira.coreos.com/browse/ODC-347](https://jira.coreos.com/browse/ODC-347)

This PR provides a script to install

the latest console with the developer perspective, and
the devconsole operator needed to enable the perspective.
The prerequisites for testing this are
export KUBECONFIG=kubeconfig file

Run the script consoledeveloper.sh
It does the following:
1. Replaces the existing openshift console with the talamer console
2. Installs the operator. (Prompts if it already exists)
3. Creates a non-admin user consoledeveloper with the password as developer with the suitable rolebinding(rolebinding being used here is self-provisioner and view)

Steps to test this

`sh consoledeveloper.sh`
oc login -u `consoledeveloper` -p `developer`
Logging in as the consoledeveloper user, you can now create a new project and do oc get csvs in the suitable namespace to see the installed operator.
Expected Output-
On the UI you can now see a consoledeveloper user under the kubeadmin option.
You can enter the username as consoledeveloper and the password as developer here.
