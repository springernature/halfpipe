package manifest

var jsonSchema = `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "$ref": "#/definitions/Manifest",
  "definitions": {
    "DeployCF": {
      "required": [
        "type"
      ],
      "properties": {
        "api": { "type": "string" },
        "deploy_artifact": { "type": "string" },
        "manifest": { "type": "string" },
        "name": { "type": "string" },
        "org": { "type": "string" },
        "password": { "type": "string" },
        "pre_promote": {
          "type": "array",
          "minItems": 0,
          "maxItems": 1,
          "items": {
            "anyOf": [
              {
                "$ref": "#/definitions/Run"
              },
              {
                "$ref": "#/definitions/DockerCompose"
              }
            ]
          }
        },
        "space": { "type": "string" },
        "type": { "enum": ["deploy-cf"]},
        "username": { "type": "string" },
        "vars": {
          "patternProperties": {
            ".*": {
              "type": "string"
            }
          },
          "type": "object"
        }
      },
      "dependencies": {
		"type": {"required": ["api", "space"]}
      },
      "additionalProperties": false,
      "type": "object"
    },
    "Docker": {
      "required": [
        "image"
      ],
      "properties": {
        "image": { "type": "string" },
        "password": { "type": "string" },
        "username": { "type": "string" }
      },
      "additionalProperties": false,
      "type": "object"
    },
    "DockerCompose": {
      "required": [
        "type"
      ],
      "properties": {
        "name": { "type": "string" },
        "save_artifacts": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "type": { "type": "string", "pattern": "docker-compose" },
        "vars": {
          "patternProperties": {
            ".*": {
              "type": "string"
            }
          },
          "type": "object"
        }
      },
      "additionalProperties": false,
      "type": "object"
    },
    "DockerPush": {
      "required": [
        "type"
      ],
      "properties": {
        "image": { "type": "string" },
        "name": { "type": "string" },
        "password": { "type": "string" },
        "type": { "type": "string", "pattern": "docker-push" },
        "username": { "type": "string" },
        "vars": {
          "patternProperties": {
            ".*": {
              "type": "string"
            }
          },
          "type": "object"
        }
      },
	  "dependencies": {
		"type": {"required": ["image"]}
      },
      "additionalProperties": false,
      "type": "object"
    },
    "Manifest": {
      "required": [
        "team",
        "tasks"
      ],
      "properties": {
        "repo": {
          "$schema": "http://json-schema.org/draft-04/schema#",
          "$ref": "#/definitions/Repo"
        },
        "slack_channel": { "type": "string", "pattern": "#.+" },
        "tasks": {
          "$schema": "http://json-schema.org/draft-04/schema#",
          "$ref": "#/definitions/Tasks"
        },
        "team": { "type": "string" },
        "trigger_interval": { "type": "string", "pattern": "\\d+h" }
      },
      "additionalProperties": false,
      "type": "object"
    },
    "Repo": {
      "properties": {
        "git_crypt_key": { "type": "string" },
        "ignored_paths": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "private_key": { "type": "string" },
        "uri": { "type": "string" },
        "watched_paths": {
          "items": {
            "type": "string"
          },
          "type": "array"
        }
      },
      "additionalProperties": false,
      "type": "object"
    },
    "Run": {
      "required": [
        "type"
      ],
      "properties": {
        "docker": {
          "$schema": "http://json-schema.org/draft-04/schema#",
          "$ref": "#/definitions/Docker"
        },
        "name": { "type": "string" },
        "save_artifacts": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "script": { "type": "string" },
        "type": { "enum": ["run"]},
        "vars": {
          "patternProperties": {
            ".*": {
              "type": "string"
            }
          },
          "type": "object"
        }
      },
      "dependencies": {
		"type": {"required": ["script", "docker"]}
      },
      "additionalProperties": false,
      "type": "object"
    },
    "Tasks": {
      "type": "array",
      "minItems": 1,
      "items": {
        "anyOf": [
          {
            "$ref": "#/definitions/Run"
          },
          {
            "$ref": "#/definitions/DockerCompose"
          },
          {
            "$ref": "#/definitions/DockerPush"
          },
          {
            "$ref": "#/definitions/DeployCF"
          }
        ]
      }
    }
  }
}
`
