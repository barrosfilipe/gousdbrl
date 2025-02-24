# 🌎 gousdbrl

> Fetch USD to BRL exchange rate from [Wise](https://wise.com/)

> **No API keys required!** :tada:

<p align="center">
  <img src="https://raw.githubusercontent.com/barrosfilipe/gousdbrl/refs/heads/main/demo.gif" />
</p>

# Install

```
go install github.com/barrosfilipe/gousdbrl@latest
```

# Run

> Before running the command, check that `$(go env GOPATH)/bin` is in your system's PATH. If not, add it to ensure globally installed Go binaries are accessible.

For `Linux/macOS`
> Add this line to your `~/.bashrc`, `~/.bash_profile`, or `~/.zshrc`:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```
> Reload your terminal configs by running
```bash
source ~/.bashrc # or source ~/.zshrc
```

Then finally run:

```
gousdbrl
```
