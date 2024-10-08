# buildpackapplifecycle

The buildpack lifecycle implements the traditional Cloud Foundry
deployment strategy.

The **Builder** downloads buildpacks and app bits, and produces a
droplet.

The **Launcher** runs the start command using a standard rootfs and
environment.

Read about the app lifecycle spec here:
https://github.com/cloudfoundry/diego-design-notes\#app-lifecycles

> \[!NOTE\]
>
> This repository should be imported as
> `code.cloudfoundry.org/buildpackapplifecycle`.

# Contributing

See the [Contributing.md](./.github/CONTRIBUTING.md) for more
information on how to contribute.

# Working Group Charter

This repository is maintained by [App Runtime
Platform](https://github.com/cloudfoundry/community/blob/main/toc/working-groups/app-runtime-platform.md)
under `Diego` area.

> \[!IMPORTANT\]
>
> Content in this file is managed by the [CI task
> `sync-readme`](https://github.com/cloudfoundry/wg-app-platform-runtime-ci/blob/c83c224ad06515ed52f51bdadf6075f56300ec93/shared/tasks/sync-readme/metadata.yml)
> and is generated by CI following a convention.
