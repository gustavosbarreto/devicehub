<h1 align="center">
  <a href="https://shellhub.io"><img src="http://docs.shellhub.io/img/logo.png" alt="ShellHub"></a>
</h1>

<h4 align="center">Centralized SSH for the edge and cloud computing.</h4>

<p align="center">
  <a href="https://github.com/shellhub-io/shellhub/actions?query=workflow%3AQA">
    <img src="https://github.com/shellhub-io/shellhub/workflows/QA/badge.svg" alt="GitHub Workflow">
  </a>
  <a href="https://gitter.im/shellhub-io/community">
    <img src="https://badges.gitter.im/shellhub-io/shellhub.svg">
  </a>
  <!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
<a href="#contributors-"><img src="https://img.shields.io/badge/all_contributors-18-orange.svg?style=flat-square"></a>
<!-- ALL-CONTRIBUTORS-BADGE:END -->
</p>

<p align="center">
  <a href="https://shellhub.io">ShellHub Hosted</a> •
  <a href="http://docs.shellhub.io">Documentation</a> •
  <a href="#contributing">Contributing</a> •
  <a href="#getting-help">Getting Help</a>
</p>

ShellHub is a modern SSH server for remotely accessing Linux devices
via command line (using any SSH client) or web-based user interface, 
designed as an alternative to _sshd_. Think ShellHub as centralized SSH
for the the edge and cloud computing.

A hosted version of ShellHub is available at https://shellhub.io/.

![Screenshot](https://github.com/shellhub-io/shellhub-io.github.io/raw/src/docs/img/screenshot.png)

## Getting started

To start using ShellHub, it is suggested that you follow the
[Getting started](https://shellhub-io.github.io/getting-started/) guide 
in the ShellHub documentation.

## Contributing

ShellHub is an open source project and we love to receive contributions from
our community. If you would like to contribute, please read
our [contributing guide](CONTRIBUTING.md).

## Getting Help

You are welcome to submit any questions, participate in discussions and request
help with any issue in our [Gitter Channel](https://gitter.im/shellhub-io/community).

## License

ShellHub is licensed under the Apache License, Version 2.0.
See [LICENSE](LICENSE.md) for the full license text.

## Development Environment Setup

First of all, you'll need to generate required keys for the services:

```
$ make keygen
```

Set the environment to development:

```
$ echo "SHELLHUB_ENV=development" >> .env.override
```

Start the whole environment:

```
$ make start
```

> Avoid to use `docker-compose` directly (unless you know what you're doing); instead use `bin/docker-compose` wrapper.

Wait for all services to be ready then create initial user by running:

```
$ ./bin/add-user <username> <password> <email>
```

Create a namespace for grouping devices and set user as owner:

```
$ TENANT_ID=00000000-0000-4000-0000-000000000000 ./bin/add-namespace <namespace> <owner>
```

> Don't change the value of `TENANT_ID`, this value is hardcoded during agent initialization in development mode.

When you open ShellHub UI for the first time, be sure to accept pending device.

See the [devscripts which can be useful for development](./devscripts).

## Authors

ShellHub was created by [O.S. Systems](https://www.ossystems.com.br).

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/gustavosbarreto"><img src="https://avatars1.githubusercontent.com/u/86747?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Luis Gustavo S. Barreto</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=gustavosbarreto" title="Code">💻</a></td>
    <td align="center"><a href="http://www.ossystems.com.br/blog"><img src="https://avatars0.githubusercontent.com/u/25278?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Otavio Salvador</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=otavio" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/leonardojoao"><img src="https://avatars1.githubusercontent.com/u/15831786?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Leonardo da Rosa Silveira João</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=leonardojoao" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/eduardoveiga"><img src="https://avatars3.githubusercontent.com/u/8249343?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Eduardo Kluwe Veiga</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=eduardoveiga" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/noreng-jg"><img src="https://avatars2.githubusercontent.com/u/25461720?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Vagner Nornberg</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=noreng-jg" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/fbertux"><img src="https://avatars1.githubusercontent.com/u/2940537?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Fabio Berton</b></sub></a><br /><a href="#platform-fbertux" title="Packaging/porting to new platform">📦</a></td>
    <td align="center"><a href="http://gomex.me"><img src="https://avatars3.githubusercontent.com/u/95132?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Rafael Gomes</b></sub></a><br /><a href="#infra-gomex" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/sixhills"><img src="https://avatars1.githubusercontent.com/u/69159771?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Mike</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=sixhills" title="Code">💻</a></td>
    <td align="center"><a href="https://inductor.me"><img src="https://avatars0.githubusercontent.com/u/20236173?v=4?s=100" width="100px;" alt=""/><br /><sub><b>inductor(Kohei)</b></sub></a><br /><a href="#infra-inductor" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a></td>
    <td align="center"><a href="https://github.com/u5surf"><img src="https://avatars1.githubusercontent.com/u/14180225?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Y.Horie</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=u5surf" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/Robinsondssantos"><img src="https://avatars3.githubusercontent.com/u/29050986?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Robinson D. S. Santos</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=Robinsondssantos" title="Code">💻</a></td>
    <td align="center"><a href="https://lbsfilm.at"><img src="https://avatars2.githubusercontent.com/u/1556271?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Lukas Bachschwell</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=s00500" title="Documentation">📖</a> <a href="https://github.com/shellhub-io/shellhub/commits?author=s00500" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/shawn111"><img src="https://avatars0.githubusercontent.com/u/346761?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Shawn</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=shawn111" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/brammp"><img src="https://avatars0.githubusercontent.com/u/52255786?v=4?s=100" width="100px;" alt=""/><br /><sub><b>brammp</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=brammp" title="Code">💻</a></td>
  </tr>
  <tr>
    <td align="center"><a href="http://xevo.com"><img src="https://avatars.githubusercontent.com/u/7035152?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Sam McKelvie</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=sammck" title="Code">💻</a></td>
    <td align="center"><a href="http://henrybarreto.dev"><img src="https://avatars.githubusercontent.com/u/23109089?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Henry Barreto</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=henrybarreto" title="Code">💻</a></td>
    <td align="center"><a href="https://www.linkedin.com/in/antony-rafael-9924511a9/"><img src="https://avatars.githubusercontent.com/u/21010565?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Antony Rafael</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=AntonyRafael" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/gessecarneiro"><img src="https://avatars.githubusercontent.com/u/55324790?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Gessé Carneiro</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=gessecarneiro" title="Code">💻</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!
