package agente

// Textos do provedor Cartesia: o passo a passo do console e os payloads
// das três client tools (mesmos nomes e parâmetros da ElevenLabs; muda só
// o envelope do formato Cartesia).

const GUIA_CARTESIA = `PROVEDOR CARTESIA — PASSO A PASSO (Standard Agent)

1. Conta e chave: play.cartesia.ai → API Keys → criar chave (sk_car_...).
   Cole em CARTESIA_API_KEY no .env — fica só no servidor; o widget recebe
   um access token de 2 minutos cunhado por /api/voz/cartesia.

2. Criar o agente: Playground → Agents → Create agent → Standard Agent
   (Recommended: a Cartesia roda STT + LLM + TTS e o loop de conversa).
   - Name: Wave
   - Instructions: o mesmo prompt de sistema da ElevenLabs (seção acima) —
     os nomes das ferramentas são idênticos.
   - Welcome Message: a mesma saudação.
   - Voice & Language: filtre as vozes por Portuguese (Brazil) e ouça antes
     de escolher; Language: Portuguese.
   - LLM: o padrão do console está bom (rápido e barato).
   - Advanced: keyterms ["WaveHub", "WaveTV", "religue", "Pix", "CPF"];
     noise suppression auto; timezone America/Sao_Paulo.
   - Background Sound: nenhum.

3. Ferramentas: aba Tools → Add tool → Client function — crie as três
   com os payloads JSON abaixo (nomes exatos, lowercase com underline).
   O diálogo já vem com um parâmetro vazio: remova-o antes de adicionar
   os seus. Depois confira que as três estão anexadas ao agente.

4. Publish: mudanças de config só valem para chamadas novas depois de
   publicar (Publish to apply this change). Rotule versões com
   version_description.

5. Copie o agent_id do agente para CARTESIA_AGENT_ID no .env (ou no campo
   do widget, com o provedor Cartesia selecionado).

6. Testar: o Playground exercita a conversa, mas client tools só disparam
   no widget — elas executam no seu navegador.`

const ESQUEMA_CARTESIA_CHAMADO = `{
  "type": "client",
  "name": "abrir_chamado_tecnico",
  "description": "Abre um chamado técnico para o cliente. Chame quando tiver o CPF e a descrição do problema — confirme o resumo com o cliente antes. O nome é opcional: com o CPF no cadastro, a ferramenta identifica o cliente sozinha.",
  "pre_tool_speech": "auto",
  "execution_mode": "immediate",
  "expects_response": true,
  "response_timeout_secs": 20,
  "parameters": {
    "type": "object",
    "properties": {
      "cpf":        { "type": "string", "description": "CPF do cliente, com ou sem pontuação" },
      "descricao":  { "type": "string", "description": "Descrição do problema em uma ou duas frases" },
      "nome":       { "type": "string", "description": "Nome completo do cliente, apenas se ele informar" },
      "problema":   { "type": "string", "enum": ["sem_internet", "sem_sinal", "lentidao", "wifi", "tv", "outro"] },
      "prioridade": { "type": "string", "enum": ["baixa", "media", "alta"] },
      "telefone":   { "type": "string" },
      "email":      { "type": "string" }
    },
    "required": ["cpf", "descricao"]
  }
}`

const ESQUEMA_CARTESIA_SEGUNDA_VIA = `{
  "type": "client",
  "name": "enviar_segunda_via_boleto",
  "description": "Envia a segunda via do boleto em aberto do cliente. Chame depois de identificar o cliente pelo CPF.",
  "pre_tool_speech": "auto",
  "execution_mode": "immediate",
  "expects_response": true,
  "response_timeout_secs": 20,
  "parameters": {
    "type": "object",
    "properties": {
      "cpf":   { "type": "string", "description": "CPF do cliente, com ou sem pontuação" },
      "email": { "type": "string", "description": "E-mail alternativo para o envio, se o cliente pedir" },
      "canal": { "type": "string", "enum": ["email", "whatsapp"] }
    },
    "required": ["cpf"]
  }
}`

const ESQUEMA_CARTESIA_RELIGUE = `{
  "type": "client",
  "name": "solicitar_religue_confirmacao",
  "description": "Registra o pedido de religue do sinal. Chame somente depois que o cliente confirmar o pagamento da fatura em aberto — sem isso a ferramenta responde comprovacao_pendente.",
  "pre_tool_speech": "auto",
  "execution_mode": "immediate",
  "expects_response": true,
  "response_timeout_secs": 20,
  "parameters": {
    "type": "object",
    "properties": {
      "cpf":                  { "type": "string", "description": "CPF do cliente, com ou sem pontuação" },
      "pagamento_confirmado": { "type": "boolean", "description": "true somente com o pagamento confirmado pelo cliente" },
      "forma_pagamento":      { "type": "string", "enum": ["pix", "cartao", "boleto", "dinheiro"] },
      "data_pagamento":       { "type": "string", "description": "Data do pagamento informada pelo cliente" },
      "comprovante":          { "type": "string", "description": "Referência do comprovante (nº de transação, horário)" }
    },
    "required": ["cpf", "pagamento_confirmado"]
  }
}`
