# mdo

Leitor descartável de Markdown para o terminal e navegador, escrito em Go.

## Uso durante o desenvolvimento

```bash
go run . README.md
```

O `mdo` valida o arquivo, escolhe uma porta local livre, abre o navegador e encerra o servidor assim que a página e os diagramas Mermaid terminam de renderizar. A página já carregada continua disponível, mas não pode ser recarregada depois que o servidor encerra.

Somente arquivos com extensão exata `.md` são aceitos.

## Recursos

- CommonMark e GitHub Flavored Markdown
- tabelas, listas de tarefas, autolinks e texto tachado
- notas de rodapé e listas de definição
- syntax highlighting
- sumário e links de títulos
- cópia de blocos de código
- diagramas Mermaid offline
- imagens relativas ao arquivo Markdown

## Build e instalação com `gobuild`

O projeto mantém um entrypoint em `./cmd` compatível com a função customizada `gobuild`:

```bash
gobuild /caminho/para/mdo
```

Estando na raiz deste repositório, também é possível executar:

```bash
gobuild "$(pwd)"
```

A função gera um binário otimizado com:

```bash
go build -ldflags="-s -w" -o mdo ./cmd
```

Depois copia o executável para `/usr/local/bin/mdo` usando `sudo`. Após a instalação:

```bash
mdo README.md
```

## Instalação por release

Os instaladores baixam o pacote adequado dos assets da última release do GitHub e instalam o executável no sistema.

### Linux

```bash
chmod +x ./install-linux.sh
./install-linux.sh
```

Instala o `mdo` em `/usr/local/bin/mdo`.

### macOS

```bash
chmod +x ./install-mac.sh
./install-mac.sh
```

Instala o `mdo` em `/usr/local/bin/mdo`.

### Windows

Em um terminal PowerShell:

```powershell
.\install-windows.ps1
```

Instala o `mdo.exe` em `$env:USERPROFILE\bin` e adiciona esse diretório ao `PATH` do usuário quando necessário. Reinicie o terminal após a primeira instalação se o `PATH` for alterado.

## Instalação global com Go

Para instalar a versão publicada do módulo globalmente:

```bash
go install github.com/egomes/mdo@latest
```

O executável será instalado em `GOBIN` ou, quando essa variável não estiver definida, em `$(go env GOPATH)/bin`. Esse diretório precisa estar no `PATH`.

Para instalar diretamente a cópia local:

```bash
go install .
```

Desde o Go 1.17, `go get` não instala mais executáveis. Em versões antigas do Go, o comando equivalente era:

```bash
go get github.com/egomes/mdo
```

Nas versões atuais, use `go install github.com/egomes/mdo@latest`; executar `go get` apenas adicionaria o módulo como dependência de outro projeto.

## Testes

```bash
go test ./...
```

Para incluir o detector de condições de corrida:

```bash
go test -race ./...
```

## Releases e binários

Ao publicar uma GitHub Release com uma tag semântica, por exemplo `v0.1.0`, o CI executa os testes e anexa à própria release os seguintes pacotes:

| Sistema | Arquitetura | Arquivo |
| --- | --- | --- |
| Linux | amd64 | `mdo-linux-amd64.tar.gz` |
| Linux | arm64 | `mdo-linux-arm64.tar.gz` |
| macOS | amd64 | `mdo-darwin-amd64.tar.gz` |
| macOS | arm64 | `mdo-darwin-arm64.tar.gz` |
| Windows | amd64 | `mdo-windows-amd64.zip` |
| Windows | arm64 | `mdo-windows-arm64.zip` |

O arquivo `checksums.txt` contém os hashes SHA-256 de todos os pacotes. Não é necessário manter um repositório separado para os binários.

Para criar uma release pela CLI do GitHub:

```bash
gh release create v0.1.0 --generate-notes
```

A publicação da release dispara o build automaticamente. Os binários também ficam disponíveis na página de releases:

```text
https://github.com/EduardoNGomes/mdo/releases
```

```mermaid
flowchart LR
    A[Markdown] --> B[mdo]
    B --> C[Navegador]
    C --> D[Servidor encerrado]
```
