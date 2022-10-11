# ProgettoSDCC

## Descrizione

In questa applicazione un certo numero di nodi concorre a popolare una pagina web con notizie di attualità. Ogni nodo scaricherà notizie relative a un tema particolare
e cercherà di pubblicare queste informazioni dinamicamente. In particolare ogni nodo vuole pubblicare una notizia per volta sulla pagina web ma tale scrittura non può avvenire in maniera concorrente. 
Per questo motivo si rende necessario l'utilizzo di un algoritmo per la mutua esclusione che garantisce che un solo nodo per volta possa scrivere la propria notizia senza creare conflitti.

## Mutua esclusione token distribuito

### Esecuzione

```bash
cd Token Distribuito
docker compose up --build
```

## Algoritmo Centralizzato

### Esecuzione

```bash
cd Autorizzazione centralizzata
docker compose up --build
```

## Lamport distribuito

### Esecuzione

```bash
cd Lamport distribuito
docker compose up --build
```

## Risultati

Una volta lanciato il comando è possibile visitare _http://localhost:8080/_ e osservare come dinamicamente il sito venga popolato da notizie e continuamente aggiornato.
Un esempio di come la pagina web apparità dopo alcuni cicli di esecuzione è il seguente.

![](https://github.com/lucaFiscariello/ProgettoSDCC/blob/d50dfb2d1db19d22de57083bbabfbcc742b5fd7e/Token%20distribuito/WebServer/webSite/template.png)
