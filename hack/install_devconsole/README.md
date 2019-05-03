## Developer console for non-admin users

You can use `consoledeveloper.sh` script available in this directory
to enable developer console for a non-admin user.

Before running the script, ensure you have set `KUBECONFIG`
environment variable pointing to the Kubernetes configuration file.

    export KUBECONFIG=/path/to/kubeconfig

Also, ensure you have the latest `oc` command (minimum version: 4.1)
available in the `PATH` environment, so that the script can use it.

You can invoke the script like this:

    ./consoledeveloper.sh

The script does the following things:

1. Replaces the existing OpenShift console with the with a temporary
   fork [talamer based console](talamer).
2. Installs the DevConsole operator.  You can see a message if it
   already exists.
3. Creates a non-admin user named `consoledeveloper` with the password
   as `developer`.  The user gets `self-provisioner` and `view`
   cluster roles.


After the script completes execution, you can see the message to login
as the `consoledeveloper` user:

    oc login -u consoledeveloper -p developer

After logged in as the `consoledeveloper` user, you can create a new
project and run `oc get csvs` in to see the installed operators.

When you open the UI, you can now see the `consoledeveloper` user
along with `kubeadmin` user.  Select `consoledeveloper` user to
proceed with login.  Enter the password as `developer`.  Now you
should be able to view the Developer perspective.

[talamer]: https://github.com/talamer/console
