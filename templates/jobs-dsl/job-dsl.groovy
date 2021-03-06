{{/* Template is defined by ProjectModel struct (project-model.go) */}}
multibranchPipelineJob('{{ .Project.Name }}') {
  description('Folder for Project {{ .Project.Name }} generated and maintained by Forjj. To update it use forjj update')
  branchSources {
{{ if eq .Project.SourceType "github" }}\
      github {
{{   if not (eq .Project.Github.ApiUrl "https://api.github.com/") }}\
          apiUri('{{ .Project.Github.ApiUrl }}')
{{   end }}\
          repoOwner('{{ .Project.Github.RepoOwner }}')
{{   if .Source.GithubUser.Name }}\
          scanCredentialsId('github-user')
{{   end }}\
          repository('{{ .Project.Github.Repo }}')
      }
{{ end }}\
{{ if eq .Project.SourceType "git" }}\
      git {
          remote('{{ .Project.Git.RemoteUrl }}')
          includes('*')
      }
{{ end }}\
  }
{{ if not (eq .Project.JenkinsfilePath "") }}\
  configure {
      it / factory {
          scriptPath('{{ .Project.JenkinsfilePath }}')
    }
  }
{{ end }}
  orphanedItemStrategy {
      discardOldItems {
          numToKeep(20)
      }
  }
}
