The halfpipe manifest is defined by a JSON Schema. Editors that support the [YAML Language Server](https://github.com/redhat-developer/yaml-language-server) can use it to validate your manifest and provide autocompletion as you type.

Add this comment to the top of your halfpipe file to enable it:

```yaml
# yaml-language-server: $schema=https://github.com/springernature/halfpipe/releases/latest/download/schema.json
team: team-name
pipeline: pipeline-name
platform: actions

triggers:
  ...
 
tasks:
  ...
```
