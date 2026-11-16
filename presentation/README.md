# BeeVibe presentation

`beevibe.pptx` — 12-slide presentation in Russian built with the same colors,
fonts and screenshots as the BeeVibe product.

## Files

- `beevibe.pptx` — editable PowerPoint deck.
- `build/pdf/beevibe.pdf` — exported proof PDF.
- `assets/fonts/` — local Unbounded, Manrope and JetBrains Mono fonts.
- `assets/shots/` — cropped screenshots from `~/Downloads/Telegram Desktop/1..6.png`.
- `build/build_presentation.py` — reproducible PPTX builder.

## Rebuild

```bash
python3 -m venv presentation/build/.venv
presentation/build/.venv/bin/pip install -q python-pptx pillow fonttools brotli
presentation/build/.venv/bin/python presentation/build/build_presentation.py
```

The builder also converts the PPTX to PDF with `soffice --headless`.
