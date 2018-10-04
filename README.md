# Forjj jenkins plugin

## Introduction

This plugin generates runnable code (deploy code) to add (create and maintain) a continuous integration component (Jenkins) to your forge.

By default, it implements latest **Jenkins 2.x** with **pipeline**, several **jenkins plugins**, **basic security rights** and a collection of scripts to enhance jenkins management from code perspective.

Mainly it offers:

- Add/Maintain new project repository to generate jobs (`projects`)
  (Jenkinsfile + JobsDSL).
- Add/Maintain Jenkins plugins/Jenkins features (`features`) (See [jenkins-install-inits(]https://github.com/forj-oss/jenkins-install-inits) and [jplugins](https://github.com/forj-oss/jplugins))
- Deploy Jenkins (master) on Docker/Swarm/UCP/Mesos/DCOS and scale it. (`applications/<instanceName>/deploy-to`)

## How to create your jenkins CI infrastructure for your organization?

Using forjj, it is quite easy.

1. Create a new jenkins instance in your Main Forjfile:

    ```yaml
    applications:
      <Instance Name>: # usually 'jenkins'
        type: ci
        driver: jenkins
    ```

2. Run `forjj update` (or `forjj create` when you create your forge, the first time.)
3. Run `forjj maintain dev` (`dev` is a deployment example name. See Forjfile for details)
4. Check at http://localhost:8080.
 
   The admin password is `MyAdmin2016` by default.
  
   To change it, call `forjj secrets set app/<Instance Name>/admin-pwd`

By default, this will create a Jenkins container started in your docker host (Forjj DooD) and become accessible at http://localhost:8080

Behind the scene, forjj-jenkins generates a collection of deployment files in a new deployment repository from the [forjj-jenkins templates files](templates).

For details about how templates works or create you own templates, read [the Jenkins source Templates section](#jenkins-source-templates)

## Jenkins source templates

All jenkins source files are generated from a collection of source templates.

By default, forjj-jenkins will use the default forjj template to build your Jenkins instance.
But you create your own forjj jenkins template to deliver your service in the way you need.

### Use of our default forjj jenkins template

The default forjj jenkins template is currently used by `forj-oss` organization to build all declared repositories that has Jenkinsfile.

It implements:

- A docker image built from `hub.docker.io/forjdevops/jenkins` [source](https://github.com/forj-oss/jenkins-ci)
- A collection of default features ([source](https://github.com/forj-oss/jenkins-install-inits))
  - Basic authentication (admin user with default password & anonymous has read access)
  - proxy setting (Set proxy from http_proxy env setting, found from the container)
  - seed-job (One job generated to populate the other collection of jobs/pipelines)
  - jenkins slave fixed port
  - pipeline
- A collection of additional features and templates to add for a dedicated deployment
- A list of predefined deployment. ie:
  - docker - To deploy to your local docker environment. (Default deployment)
  - ucp - To deploy to a UCP system. **NOT AVAILABLE**
  - marathon - To deploy to dcos/mesos marathon. **NOT AVAILABLE**

The template is configured by your `Forjfile` and `forjj secrets`.

From the Forjfile, you can add more plugins, configure them and configure Jenkins. Forjj will transmit all such information to properly configure Jenkins
as requested.

For example, by default, jenkins is configured with a jenkins admin account and a default password.
You can update it with `forjj secrets edit app/jenkins/admin-pwd`


### Create your own forjj jenkins template

If you want to manage your Jenkins instance differently than what the default template do, we will need to create your own forjj jenkins template.

This section explains baby steps to start creating it. For template details, check our [template documentation] (TODO)

NOTE: Replace any `<*>` to match your need.

1. In your infra repository, creates  `<infraRepo>/apps/ci/<instanceName>/templates/<templateName>`

    Ex:

    ```bash
    cd ~/forjj/infra
    mkdir -p apps/ci/jenkins/template/jenkins-aws
    ```

2. Update your master Forjfile and add `source-template: <templateName>` under `applications/<instanceName>`

    Ex:
  
    ```yaml
    applications:
      jenkins:
        type: ci
        source-template: jenkins-aws
    ```
  
3. Optionnaly, set a deployment name with `deploy-to` in your Forjfile. By default, forjj-jenkins set it to `docker`

    Depending on the environ you want forjj to deliver (dev, test, production) the deployment technology can be completely different, but must be supported
    by your team.

    Example:
    If your team managing Jenkins want to test a new feature or a new plugin, your factory help them to deliver such service under a docker container locally
    on their workstation, while on production, you use `aws` or `kubernetes`, or any kind of infrastructure to host Jenkins service

    So you can have a `docker` deployment type and `aws` deployment type

    So, the Jenkins development team while running `forjj maintain` will use `docker` by default, but
    the Jenkins software factory pipepine will deliver with `forjj maintain production` against the `aws` deployment, which will call ansible, kubctl, ... to create and maintain your production instance of Jenkins

    Ex: In production Forjfile
  
    ```yaml
    applications:
      jenkins:
        type: ci
        deploy-to: aws-ansible
    ```

3. Create the `template.yaml` which defines collection of files, default features and build/deploy commands.

    forjj-jenkins thanks to `deploy-to` can manage differents list of files and feae

    At least, a `templates.yaml` must exist with one deplo

    ```yaml
    ---
    sources: # List all source or template files
      common: # All files described under this section will be generated by forjj-jenkins
        # Add one of more:
        # <fileTitle>: { template: , source: , chmod: , tags: }
        my 1st file: # First use case: Copy a file from template directory.
          source: relative/path/to/my/file/to/copy # template relative file path which needs to be copied
          chmod:  0755 # optional. support chmod octal representation only. valid for `source` or `template` file type.
          if:        # Optional. A template string. If is true if the result of this template return a non null string. 
                     # valid for `source` or `template` file type.
        my 2nd file: # Seconf use case: Generate a file from template directory and deployment template data (<plugin>.yaml explained later).
          template: relative/path/to/my/template/file # template relative file path to a source template file used to generate this file.
                     # A template file must contains at least {{ }} tag for template replacement. You can change the tag with `tags`.
          tags: # string which describes a template tag to use. By default, it uses {{}}. The string is cut in 2 sub strings of same size.
                # The first one is the opened tag like '{{'
                # The second one the closed tag like '}}'
      deploy: # Depending on `deploy-to` parameter of your Forjfile
        docker: # This is the default `deploy-to`. Same as Common section.
          [...]
        <DeployOption>: # same as common section.
          [...]
    features:
      common: # Following is basic list of features or plugins to install. This will be the default list to apply on all deployments
      - "feature:jenkins-init"
      - "feature:seed-job"
      - "feature:tcp-slave-agent-port"
      - "feature:jenkins-pipeline"
      - "feature:credentials"
      - "{{ if .Source.ProjectsHasSource \"github\" }}feature:multibranch-github-pipeline{{ end }}"
      - "{{ if .Source.ProjectsHasSource \"bitbucket\" }}plugin:cloudbees-bitbucket-branch-source{{ end }}"
      deploy:
        docker: # List of plugins specifically for the default docker deployment.
          [...]
        <DeployOption>: # List of plugins specifically this other deployment.
          [...]
    run_deploy: # Called at maintain request
      docker:
        run: # string representing the command to run
        environment: # Optional
          <environmentVariableName>:
            if: # Optional. A template string. If is true if the result of this template return a non null string. if false the environment variable is not set.
            value: # a string or a template string
        files: # Optional
          <PathToFile>:
            content: "{{ index .Creds \"app-jenkins-feature-credentials-json\" }}" # a string or a template string
            if: # Optional. A template string. If is true if the result of this template return a non null string
            remove-when-done: true # default is false. if true, the file created will be removed at the end of the script run.
      <DeployOption>: # same as docker for each deployment.
        [...]
    run_build: # Same as run_deploy but called at create/update request
      [...]
    ```

### Running `run_deploy` scripts

when forjj-jenkins call a script described by `run_deploy` or `run_build`, a shell command is started with:

- The list of environment as described by `environment`
- A list of files as described by `files`, created before and removed after if requested.
- a collection of pre-defined environment variable sent by `forjj-jenkins`
  - DOCKER_DOOD        : docker run helper given by forjj to start a container in DooD mode.
  - DOCKER_DOOD_BECOME : docker run helper given by forjj to start a container with a different UID/GID.
  - DOCKER_DOOD_PROXY  : docker run helper given by forjj to start a container with a proxy setup.
  - DOOD_SOURCE        : docker run helper given by forjj to start a container with options to mount source/deploy/workspace path

  - DOOD_SRC           : Obsolete. Use DOOD_SOURCE. Host path to the source code.
  - DOOD_DEPLOY        : Obsolete. Use DOOD_SOURCE. Host path to the deployement code.

  - GID                : Group ID for the current user.
  - UID                : User ID for the current user.
  - LOGNAME            : User name of the current user.
  - PATH               : Shell path
  - TERM               : Terminal string
  - HOSTNAME           : Host or container name
  - http_proxy         : Proxy information if set
  - https_proxy        : Proxy information if set
  - no_proxy           : Proxy information if set
  - SELF_SRC           : Container Path to the source directory tree.
  - SELF_DEPLOY        : Container Path to the deployments directory tree

NOTE: when forjj-jenkins run a command, it displays each environment variable used to run the command.

## Forjfile plugin options

forjj-jenkins has several objects and that can be defined in your Forjfile.

For details, read [jenkins.yaml](jenkins.yaml)

**NOTE**: In short future, forjj will show up all plugins options with something like `forjj show`

## github upstream with pull-request flow setting.

The github integration will update your `<infra>/ci/<instance Name>` with the following code.

- `pipeline github` feature
- 1 Jobs DSL for each project identified under `<deploy Repo>/<instance Name>/jobs-dsl/*.groovy`

### Other SCM

Currently this jenkins Forjj plugin do not have any other upstream integration.
But this CI orchestrator has been designed to easily add a new one, like gitlab or other flows.
If you want to add you SCM/Jenkins integration, consider contribution to this repository.

For details on contribution, see CONTRIBUTION.md

## Contribution

Feel free to contribute and create a pull request or issue.

For details on contribution, see CONTRIBUTION.md

## Todo

Following list is not exhaustive and is given as an example. Check list of enhancement or issues in [forjj-jenkins repository](https://github.com/forj-oss/forjj-jenkins/issues)

- `applications/<instanceName>/source-templates` could define a git repo, so that we can share the template more easily.

Forj team
