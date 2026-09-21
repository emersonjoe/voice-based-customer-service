package chamados

import "github.com/emersonjoe/trilha"

// Kind muda a natureza deste ramo: a escrita dele (o POST de status, chamado
// por formulários do painel) herda a conferência de CSRF de página em vez do
// comportamento de API aberta.
var Kind = trilha.KindPage
