team: halfpipe-team
pipeline: pipeline-name
platform: actions

triggers:
  - type: git
    watched_paths:
      - e2e/actions/deploy-katee

tasks:
  - type: deploy-katee
    name: deploy to katee
    applicationName: BLAHBLAH
    buildVersion: 0.0.${{ github.run_number }}
    # optional
    imageScanSeverity: SKIP
    applicationRoot: e2e/actions/deploy-katee
    environment: live
    url: https://ee-test-actions.public.springernature.app
    vaultEnvVars: |
      springernature/data/engineering-enablement/cloudfoundry-test user | VERY_SECRET
