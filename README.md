
# httpatos

Biblioteca HTTP minimalista para Go, escrita do zero para meu [patos-psel](github.com/gdefaria/patos-psel). Suporta roteamento por método, parâmetros de rota dinâmicos, parsing de JSON no body e controle total sobre a resposta.
> Biblioteca escrita para fins de aprendizagem, não utilize em produção.

---

## Instalação

```bash
go get github.com/gdefaria/httpatos
```
---

## Conceitos básicos

### Router
Toda aplicação começa com a criação de um router via `httpatos.Router()`. É o responsável por registrar rotas e iniciar o servidor TCP.

```go
package main

import "github.com/gdefaria/httpatos"

func main() {
    app := httpatos.Router()
    // registre suas rotas aqui...
    app.Listen(8080)
}
```

`app.Listen(porta)` é uma chamada bloqueante. Abre um socket TCP e fica aguardando conexões. Cada conexão é tratada em uma goroutine separada.

### Context

Cada handler recebe um `*httpatos.Context`, que contém duas partes:

- `ctx.Request`: informações da requisição recebida
- `ctx.Response`: construtor da resposta a ser enviada

```go
app.Get("/ping", func(ctx *httpatos.Context) {
    ctx.Response.Text("pong").Send()
})
```

> **Importante:** todo handler deve terminar com `.Send()` na resposta. Sem ele, o servidor ficará aguardando indefinidamente.

---

## Rotas GET

Use `app.Get(path, handler)` para registrar uma rota que responde a requisições `GET`.

```go
app.Get("/hello", func(ctx *httpatos.Context) {
    ctx.Response.Text("Olá, mundo!").Send()
})
```

---

## Rotas POST

Use `app.Post(path, handler)` para registrar uma rota `POST`.

```go
app.Post("/echo", func(ctx *httpatos.Context) {
    ctx.Response.Text("Recebi sua requisição!").Send()
})
```

---

## Parâmetros de rota

Parâmetros dinâmicos são declarados com `:nome` no path. Eles ficam disponíveis em `ctx.Request.Params`, que é um `map[string]string`.

```go
app.Get("/users/:id", func(ctx *httpatos.Context) {
    id := ctx.Request.Params["id"]
    ctx.Response.Text("Usuário: " + id).Send()
})
```

Uma requisição para `/users/42` vai popular `Params["id"]` com `"42"`.

É possível combinar múltiplos parâmetros no mesmo path:

```go
app.Get("/posts/:postId/comments/:commentId", func(ctx *httpatos.Context) {
    postId    := ctx.Request.Params["postId"]
    commentId := ctx.Request.Params["commentId"]
    ctx.Response.Text("Post " + postId + ", comentário " + commentId).Send()
})
```

**Prioridade:** rotas literais têm precedência sobre rotas com parâmetro. Isso significa que `/users/profile` e `/users/:id` podem coexistir — `/users/profile` sempre vai bater na rota literal.

---

## Lendo o body como JSON

Para rotas que recebem JSON no body, use `ctx.Request.JSON(&destino)`. O método valida o `Content-Type: application/json` e desserializa o body na struct fornecida. Ele retorna `true` em caso de sucesso e `false` caso o content-type seja inválido ou o JSON esteja malformado.

```go
type CreateUserPayload struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

app.Post("/users", func(ctx *httpatos.Context) {
    var payload CreateUserPayload

    if !ctx.Request.JSON(&payload) {
        ctx.Response.Status(400).Text("JSON inválido").Send()
        return
    }

    ctx.Response.Status(201).Text("Usuário criado: " + payload.Name).Send()
})
```

A requisição correspondente deve incluir o header `Content-Type: application/json`:

```
POST /users HTTP/1.1
Content-Type: application/json

{"name": "Maria", "email": "maria@exemplo.com"}
```

---

## Resposta em texto simples

Use `.Text(string)` para enviar uma resposta com body em texto puro. O header `Content-Type: text/plain` é setado automaticamente.

```go
ctx.Response.Text("tudo certo").Send()
```

---

## Resposta em JSON

Use `.Json(any)` para serializar qualquer valor Go como JSON. O header `Content-Type: application/json` é setado automaticamente.

```go
type UserResponse struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

app.Get("/users/:id", func(ctx *httpatos.Context) {
    user := UserResponse{ID: 1, Name: "Maria"}
    ctx.Response.Json(user).Send()
})
```

---

## Status HTTP

Use `.Status(código)` para definir o código de status da resposta. O método é encadeável com `.Text()`, `.Json()` ou chamado sozinho antes de `.Send()`.

```go
// 201 Created
ctx.Response.Status(201).Json(novoRecurso).Send()

// 204 No Content
ctx.Response.Status(204).Send()

// 404 Not Found
ctx.Response.Status(404).Text("não encontrado").Send()
```

Se `.Status()` não for chamado, o padrão é **200 OK**.

Os status texts são gerados automaticamente para os códigos mais comuns (100, 200, 201, 204, 301, 302, 400, 401, 403, 404, 405, 409, 410, 422, 429, 500, 501, 502, 503, 504, entre outros). Para códigos desconhecidos, o texto será `"Unknown"`.

---

## Headers personalizados

O campo `ctx.Response.Headers` é um `map[string]string` que pode ser modificado diretamente antes de `.Send()`.

```go
app.Get("/download", func(ctx *httpatos.Context) {
    ctx.Response.Headers["Content-Disposition"] = "attachment; filename=\"arquivo.txt\""
    ctx.Response.Headers["X-App-Version"] = "1.0.0"
    ctx.Response.Text("conteúdo do arquivo").Send()
})
```

> **Atenção:** `Content-Type` e `Content-Length` são gerenciados pela biblioteca (`.Text()`, `.Json()` e `.Send()` os preenchem automaticamente). Setar `Content-Length` manualmente será sobrescrito pelo `.Send()`.

---

## Body manual

Além de `.Text()` e `.Json()`, é possível atribuir o body diretamente como `[]byte` via `ctx.Response.Body`. Isso é útil para formatos arbitrários (HTML, binário, etc.).

```go
app.Get("/pagina", func(ctx *httpatos.Context) {
    ctx.Response.Headers["Content-Type"] = "text/html"
    ctx.Response.Body = []byte("<h1>Olá, mundo!</h1>")
    ctx.Response.Send()
})
```

---

## Acessando headers da requisição

Os headers da requisição ficam em `ctx.Request.Headers`, também um `map[string]string`. As chaves são normalizadas para letras minúsculas (conforme a RFC 9110).

```go
app.Get("/info", func(ctx *httpatos.Context) {
    agente := ctx.Request.Headers["user-agent"]
    ctx.Response.Text("Seu agente: " + agente).Send()
})
```

---

## Tratamento de erros internos

Se um handler entrar em pânico (`panic`), o servidor recupera automaticamente via `recover()` e retorna **500 Internal Server Error**, evitando que o processo caia. O erro é impresso no stdout para facilitar o diagnóstico.

---

## Exemplo completo

```go
package main

import (
    "fmt"
    "github.com/gdefaria/httpatos"
)

type LoginPayload struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token string `json:"token"`
}

func main() {
    app := httpatos.Router()

    // rota simples
    app.Get("/ping", func(ctx *httpatos.Context) {
        ctx.Response.Text("pong").Send()
    })

    // parâmetro de rota
    app.Get("/users/:id", func(ctx *httpatos.Context) {
        id := ctx.Request.Params["id"]
        ctx.Response.Json(map[string]string{"id": id}).Send()
    })

    // POST com JSON no body
    app.Post("/login", func(ctx *httpatos.Context) {
        var payload LoginPayload
        if !ctx.Request.JSON(&payload) {
            ctx.Response.Status(400).Text("body inválido").Send()
            return
        }

        if payload.Username != "admin" || payload.Password != "1234" {
            ctx.Response.Status(401).Text("credenciais incorretas").Send()
            return
        }

        ctx.Response.Status(200).Json(LoginResponse{Token: "abc123"}).Send()
    })

    // header personalizado
    app.Get("/download", func(ctx *httpatos.Context) {
        ctx.Response.Headers["Content-Disposition"] = "attachment; filename=\"dados.txt\""
        ctx.Response.Text("conteúdo do arquivo").Send()
    })

    fmt.Println("Servidor ouvindo em :8080")
    app.Listen(8080)
}
```

---

## Limitações
- O tamanho máximo do body de entrada é **5 MB** por conexão.
- Não há suporte a middlewares, CORS, ou TLS nativamente — funcionalidades que precisariam ser implementadas na camada da aplicação.
- Cada conexão aceita apenas uma requisição (sem keep-alive/HTTP pipelining).

