---
page_title: "jumpserver_user_group Resource"
description: |-
  Read an existing JumpServer object for read-only import-plan evaluation.
---
# jumpserver_user_group

Read-only evaluation resource. Existing objects can be described with a resource block and an `import` block, then inspected with `terraform plan`. No create/update/delete operations are supported. Do not run apply in this evaluation release.

Use the provider schema (`terraform providers schema -json`) for the exact attribute definitions. All configuration must match the existing object. A no-change result covers only modeled fields; it does not establish full permission fidelity or successful adoption into state.
