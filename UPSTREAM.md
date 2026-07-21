# Sync com o repositório oficial

Este fork ByteRush (`wesleymaia86/google-maps-scraper`) parte do `gosom/google-maps-scraper` (via branch Playwright de `atj393`) e adiciona patches próprios.

## Remotes sugeridos

```bash
git remote add upstream https://github.com/gosom/google-maps-scraper.git
git remote add deploy https://github.com/wesleymaia86/google-maps-scraper.git
```

## Fluxo de sync

1. `git fetch upstream`
2. Criar branch `sync/upstream-YYYYMMDD` a partir de `main`
3. `git merge upstream/main` (ou rebase, se preferir histórico linear)
4. Resolver conflitos priorizando os patches ByteRush abaixo
5. Testar localmente (ou deploy de staging): scrape curto, geocode, download Excel, exclude-job
6. `git push deploy sync/upstream-YYYYMMDD:main` (ou PR + merge)
7. Redeploy no Coolify (`u9s0mj80wyu966bva01ef4u1`) com force rebuild

## Patches ByteRush a revalidar após sync

| Patch | Onde |
|-------|------|
| Playwright driver 1.60.0 (CDN morto) | `Dockerfile` / install driver |
| UTF-8 BOM no download CSV | `web/web.go` → `download()` |
| Geocode latlng.work | `web/geocode.go`, env `LATLNG_API_KEY` |
| Excluir leads de job anterior | `web/exclude.go`, `runner/webrunner/excludewriter.go`, `ExcludeJobID` |
| Filename = nome do job | `web/filename.go` + `download()` |
| UI pt-BR ByteRush | `web/static/templates/*`, `web/static/css/main.css` |
| Defaults BR (`lang=pt`) | `web/web.go` → `index()` |

## Variáveis de ambiente (Coolify)

- `LATLNG_API_KEY` — chave [latlng.work](https://latlng.work) (nunca commitada)
- `DISABLE_TELEMETRY=1` — recomendado

## Persistência

Monte volume persistente em `/gmapsdata` no Coolify para jobs/CSVs sobreviverem a redeploys.
