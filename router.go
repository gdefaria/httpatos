package httpatos

import (
	"fmt"
	"strings"
)

var ALLOWED_METHODS = [...]string{"GET", "POST"}

// callback que recebe o contexto da requisição
type routeCallback func(context *Context)

type Context struct {
	Request  *httpRequest
	Response *httpResponse
}

/*
cada nó representa um segmento do path. por exemplo, a rota
GET /user/:id/posts gera três nós encadeados: "user" -> ":id" -> "posts"

filhos literais ficam em children["segmento"]; filhos paramétricos
ficam em paramChild (apenas um por nível). na busca, literais têm
prioridade e só buscamos em paramChild se children[segmento] não existir

dessa forma, é possível suportar, ao mesmo tempo, rotas como
/user/:id e /user/profile
*/
type trieNode struct {
	children   map[string]*trieNode // subrotas literais
	paramChild *trieNode            // subrota com o parâmetro
	paramName  string               // nome do parâmetro para popular o map de params
	callback   routeCallback        // callback passado pelo usuário para lidar com a rota
}

type router struct {
	// uma árvore para cada método
	routeTree map[string]*trieNode
}

func Router() *router {
	r := router{
		routeTree: make(map[string]*trieNode),
	}

	// inicializando uma árvore para cada método
	for _, method := range ALLOWED_METHODS {
		r.routeTree[method] = &trieNode{
			children: make(map[string]*trieNode),
		}
	}

	return &r
}

/*
função responsável por mapear a rota+parâmetros na árvore de rotas do método.
tenha certeza que entendeu a estrutura de trieNode antes de ler essa função.
você pode me perguntar pessoalmente caso não entenda.
*/
func (r *router) registerRoute(method string, path string, c routeCallback) {
	segments := strings.Split(path, "/")

	/*
		loop para inicializar o caminho na árvore de rotas caso ainda não exista.
		começamos no 1, pois o primeiro segmento do split é sempre uma string vazia (""),
		que representa a raiz `/` e já foi inicializada.
	*/
	node := r.routeTree[method] // representa a raíz

	for i := 1; i < len(segments); i++ {
		segment := segments[i]

		if len(segment) == 0 {
			// é uma barra no final do path, como `/path/`, que deve dar fallback para `/path`
			continue
		}

		isParam := segment[0] == ':'

		// verifica se já existe um segmento caso seja rota literal
		if !isParam && node.children[segment] != nil {
			// continua progredindo
			node = node.children[segment]
			continue
		}

		// verifica se já existe um segmento caso seja parâmetro
		if isParam && node.paramName == segment[1:] {
			// continua progredindo
			node = node.paramChild
			continue
		}

		// não há mais rotas na árvore, mas ainda há segmentos. precisamos começar a criá-las
		newNode := &trieNode{
			children: make(map[string]*trieNode),
		}

		if isParam {
			// caso seja parâmetro
			node.paramChild = newNode
			node.paramName = segment[1:]
		} else {
			// caso seja rota literal
			node.children[segment] = newNode
		}

		node = newNode
	}

	// agora que o caminho na árvore foi inicializado, podemos editar o node em questão
	node.callback = c
}

/*
função responsável por encontrar o callback e os parâmetros de uma rota crua
*/
func (r *router) matchRoute(method string, path string) (routeCallback, map[string]string) {
	params := make(map[string]string)
	segments := strings.Split(path, "/")
	node := r.routeTree[method]

	for _, segment := range segments[1:] {
		if len(segment) == 0 {
			// é apenas uma barra `/`. então deve dar fallback para o path sem a barra
			continue
		}

		// buscando um node literal (prioridade)
		literalChild := node.children[segment]

		if literalChild != nil {
			node = literalChild
			continue
		}

		// buscando por parâmetro
		paramName := node.paramName

		if paramName != "" {
			params[paramName] = segment
			node = node.paramChild
			continue
		}

		// rota não encontrada
		return nil, nil
	}

	return node.callback, params
}

/* MÉTODOS */
func (r *router) Get(path string, c routeCallback) {
	r.registerRoute("GET", path, c)
}

func (r *router) Post(path string, c routeCallback) {
	r.registerRoute("POST", path, c)
}

// função principal que recebe uma requisição crua e a direciona
func (r *router) receiveRawRequest(rawRequest []byte) (rawResponse []byte) {
	request, err := decodeHttpRequest(string(rawRequest))

	if err != nil {
		// TODO: Handle bad request
		fmt.Println("Bad Request")
		return
	}

	// verificando se o método é permitido
	isMethodAllowed := false

	// go não tem uma função built-in para verificar ocorrência em array
	for _, method := range ALLOWED_METHODS {
		if request.Method == method {
			isMethodAllowed = true
		}
	}

	if !isMethodAllowed {
		// TODO: Handle bad method
		fmt.Println("Bad method:", request.Method)
		return
	}

	// buscando o callback + parâmetros
	callback, params := r.matchRoute(request.Method, request.Path)

	if callback == nil {
		// TODO: Handle route not found
		fmt.Println("Route not found:", request.Path)
		return
	}

	// construindo o contexto
	c := &Context{
		Request:  request,
		Response: NewHTTPResponse(),
	}

	request.Params = params

	defer func() {
		if p := recover(); p != nil {
			// TODO: Handle internal server error
			fmt.Printf("[Internal Server Error]: %v", p)
		}
	}()

	callback(c)
	for !c.Response.Ready {
	} // espera até que Request.Ready == true

	return c.Response.serialize()
}

// cria um novo listener
func (r *router) Listen(port int) {
	listener := newListener(port, r.receiveRawRequest)
	listener.listen()
}
