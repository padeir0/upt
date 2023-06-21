# UFF Portugol Tools

Work in Progress:

 - [x] Spec
 - [x] Lexer
 - [x] Parser
 - [x] Name Resolution
 - [x] TypeChecking
 - [x] C Generation
 - [ ] IR Generation
 - [ ] Termination checking
 - [ ] IR Interpreter
 - [ ] Debugger
 - [ ] LSP server and client
 - [ ] Highlighting

Implementa ferramentas para o Portugol da UFF.
O resto desse documento é a especificação da linguagem.

## Sumário

1. [Elementos Lexicos](#elementoslexicos)
    1.  [Espaços](#espaços)
    2.  [Identificadores](#identificadores)
    3.  [Palavras Chave](#palavraschave)
    4.  [Operadores e Pontuação](#operadoresepontuacao)
    5.  [Literais](#literais)
    6.  [Comentários](#comentarios)
2. [Elementos Gramaticais](#elementosgramaticais)
3. [Funções embutidas](#funcoesembutidas)
4. [Tipos](#tipos)

## Elementos Lexicos <a name="elementoslexicos"/>

### Espaços <a name="espacos"/>

Espaços, quebra de linha e tabs tem como seu unico valor semantico
a separação de tokens (elementos lexicos), de resto,
eles são desconsiderados exceto dentro de strings.

### Idenficadores <a name="identificadores"/>

```
Ident := letra {letra | digito}.
```

Identificadores são simbolos textuais que começam com uma letra
e podem conter zero ou mais letras ou digitos após essa primeira.

### Palavras chaves <a name="palavraschave"/>

Esses identificadores são especiais e não podem ser usados
como nomes de simbolos na linguagem.

```
inteiro    real      caractere
para       enquanto  se    senao
imprima    leia      ou    e    nao
```

### Operadores e pontuação <a name="operadoresepontuacao"/>

```
(   )   {   }   ,   =
==  !=  >   >=  <   <=
+   -   /   *   %   ;
"   '
```

Esses simbolos são especiais e denotam operações ou separação
entre elementos gramaticais.

### Literais <a name="literais"/>

```
mensagem := '"' {ascii} '"'.
literalCaractere := '\'' ascii '\''.
literalInteiro := digito {digito}.
literalReal := digito {digito} '.' {digito}.
```

Literais são elementos lexicos mais complexos que denotam valores
basicos que podem ser usados na linguagem.

Note que o espaço tem valor semantico aqui: `9 9` são dois números
inteiros, diferente de `99`, assim como `0. 1` é um número real seguido por um inteiro
e `0.1` é apenas um número real.

A diferenciação entre `literalInteiro` e `literalReal` requer lookahead arbitrario,
mas pode ser resolvida com um truque no lexer.

### Comentários <a name="comentarios"/>

A linguagem segue o estilo de comentários de C, ou seja,
existem dois tipos de comentários:

 - Os que começam com `//` e terminam com uma quebra de linha.
 - Os que começam com `/*` e terminam com `*/`

Ambos são desconsiderados durante a compilação.

## Elementos Gramaticais <a name="elementosgramaticais"/>

A gramatica a seguir está em Wirth Syntax Notation, onde nomes começando com
letra maiuscula representam simbolos gramaticais, e nomes começando com
letra minuscula representam simbolos lexicos.

```ebnf
Portugol := {Funcao}.

Funcao := [tipo] ident '(' [ArgList] ')' Bloco.
ArgList := Arg {',' Arg}.
Arg := tipo ident.

Bloco := '{' {Comando} '}'.

Comando := Atrib term
         | VarDecl term
         | Expr term
         | Leia term
         | Imprima term
         | Se
         | Enquanto
         | Para
         | Retorne term.

Retorne := 'retorne' Expr.

Leia := 'leia' '(' ident ')'.
Imprima := 'imprima' '(' ImpArg ')'.
ImpArg := mensagem | Expr.

Atrib := ident "=" Expr.
VarDecl := tipo Idlist.

Se := 'se' '(' Expr ')' Bloco [Senao].
Senao := 'senao' Bloco.

Enquanto := 'enquanto' '(' Expr ')' Bloco.

Para := 'para' '(' [Atrib] term Expr term Atrib ')' Bloco.

ExprList := Expr {',' Expr}.
Expr := AndExpr {'ou' AndExpr}.
AndExpr := CondExpr {'e' CondExpr}.
CondExpr := AddExpr {compOp AddExpr}.
compOp := '==' | '!=' | '>' | '>=' | '<' | '<='.
AddExpr := MultExpr {addOp MultExpr}.
addOp := '+' | '-'
MultExpr := Unary {multOp Unary}.
multOp := '*' | '/' | '%'.
Unary := [unaryOp] Termo [Call].
unaryOp := '-' | 'nao'.
Call := '(' [ExprList] ')' 
Termo := literalInteiro
       | literalReal
       | literalCaracter
       | ident
       | '(' Expr ')'.

IdList := ident {',' ident}.
ident := letra {letra | digito}.

letra := 'a'|'b'|...|'z'|'A'|'B'|...|'Z'.
digito := '0'|'1'|...|'8'|'9'.

tipos := 'inteiro' | 'real' | 'caracter'.
term := ';'.
mensagem := '"' {ascii} '"'.

literalCaractere := '\'' ascii '\''.
literalInteiro := digito {digito}.
literalReal := digito {digito} '.' {digito}.
```

## Funcões Embutidas <a name="funcoesembutidas"/>

 - `raiz`: raiz quadrada: `raiz(4) == 2`
 - `expo`: potenciação/expoente: `expo(2, 2) == 4`

Note que `leia` e `imprima` não são funções em si, já que
seus tipos podem ser polimorficos e suas semanticas diferem de 
funções normais.

## Tipos <a name="tipos"/>

 - `inteiro` -> inteiro com sinal (`int`)
 - `real` -> ponto flutuante (`double`)
 - `caractere` -> inteiro de 8 bits com sinal (`char`)