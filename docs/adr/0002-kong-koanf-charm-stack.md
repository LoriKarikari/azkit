# Kong, Koanf, and Charm stack for CLI architecture

We will use Kong for command and flag parsing, Koanf for layered configuration, and the Charm stack for terminal UX. Flags override environment variables, which override config files, which override defaults. Bubble Tea, Huh, and Lip Gloss support polished interactive flows. We are avoiding Cobra to reduce boilerplate and keep commands typed, lightweight, and testable.
