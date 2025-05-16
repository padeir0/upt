# Compilador de Portugol da UFF

Implementa um compilador da linguagem de algoritmos
para C. Não, não é só um monte de macros, esse compilador
verifica o seu código e te dá erros em português.
Infelizmente, verificações mais robustas vistas quando
compilamos um programa em C com `-Wall -Wextra -Wstrict-conversion`
não foram implementadas aqui, muito provavelmente nem vão ser.

Para usar esse compilador, basta ir na aba _Releases_ do github
e baixar o binário compilado. Se está usando isso no sistema Ubuntu,
como nos computadores da UFF, basta renomear o binário do jeito que preferir
e colocar ele na pasta `~/bin`, daí você pode usar ele direto:
 
```bash
upt exemplos/muitosOlas.uffp
```

Veja que tem uma leve diferença na linguagem:
para você rodar o programa, é necessário uma função `entrada`
que serve como a `main`. Ela não pode ter um nome arbitrário
como é dado em algoritmos. Uma especificação
informal e provavelmente incompleta se encontra em [spec.md](./spec.md)

Algumas opções adicionais existem e talvez sejam úteis
para entender como o compilador lê o código, basta
chamar `upt --help`.

Se achar qualquer erro, pode abrir um issue,
me contatar pessoalmente,
ou me mandar um alô em `iureartur arroba id ponto uff ponto br`.

Peace and long life.
