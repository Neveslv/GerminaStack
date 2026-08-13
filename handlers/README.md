# handlers

Camada HTTP da aplicacao. Os handlers validam payloads, leem a identidade da requisicao e traduzem erros para status HTTP; contratos de persistencia e tipos de entrada ficam em `domain/`, nao em `database/`.
