# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**bormeparser** is a Python 3 library for parsing BORME (Boletín Oficial del Registro Mercantil) files - the official Spanish Mercantile Register bulletin. It extracts and structures business data from BORME PDFs which contain company registration information not available in XML/JSON formats.

**Status**: This project is being converted from Python to Go.

## Commands

```bash
# Installation
pip install -e .                    # Development install
python setup.py install             # Standard install

# Testing
python setup.py test                # Run all tests
python -m unittest bormeparser.tests.test_borme
python -m unittest bormeparser.tests.test_bormeparser
python -m unittest bormeparser.tests.test_bormeregex

# Documentation
cd docs && make html               # Build HTML docs
make -e SPHINXOPTS="-D language='en'" html  # English docs

# Coverage
coverage run --source=bormeparser setup.py test
coveralls --service=github
```

## Architecture

### Core Data Flow

```
parse() [parser.py] → Router → Backend (pypdf2 or lxml)
                                      ↓
                              BormeAnuncio objects
                                      ↓
                              JSON/XML export
```

### Key Components

| Component | Purpose |
|-----------|---------|
| `parser.py` | Main entry point, routes to appropriate backend based on section |
| `borme.py` | Core models: Borme, BormeAnuncio, BormeActoCargo, BormeActoTexto |
| `backends/base.py` | BormeAParserBackend base class for Section A parsers |
| `backends/pypdf2/parser.py` | PyPDF2-based PDF parser for Section A |
| `backends/seccion_c/lxml/parser.py` | Lxml-based parser for Section C |

### Backends

- **Section A** (PyPDF2): Company creation/dissolution announcements
- **Section C** (Lxml): Other announcements in HTML format

## Go Conversion Notes

When implementing in Go:

- PDF parsing: Consider `rsc/pdf` or `junglejp/go-pdf`
- XML processing: Use standard `encoding/xml` or `xmlpath` package
- HTTP requests: Standard `net/http`
- Structure parsed data into idiomatic Go types matching Python models
- Maintain the backend abstraction pattern for different parser implementations
