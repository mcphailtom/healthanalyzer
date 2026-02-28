# HealthAnalyzer

HealthAnalyzer is a personal health tracking tool that uses AI to spot patterns and trends across different areas of your daily life.

## How It Works

You provide information about your day in plain language -- what you ate, how you slept, what activity you did. The app analyses each area separately, then looks across them for connections. For example, it might notice that your sleep quality tends to drop after days with high caffeine intake, or that your energy levels improve during weeks where you're more active.

You don't need to fill in forms or use a specific format. Just describe your day naturally and the app takes care of the rest.

## What It Learns

The app keeps a history of your entries and its own analysis. Each time you submit new information, it looks back at what's come before -- both recent days and older entries that are similar to what you're describing now. This means it gets more useful over time as it builds a picture of your patterns.

Over longer periods, the app produces weekly and monthly summaries so that trends over weeks and months are not lost as daily details accumulate.

## Your Data, Your Machine

Everything runs locally. Your health data is stored in a single file on your computer. You can choose to use cloud-based AI services (OpenAI or Anthropic) for the analysis, or run a fully local AI model so that nothing leaves your machine. Even with cloud AI, only the text you submit is sent for analysis -- nothing is stored externally.

## How You Use It

The app has a web interface you open in your browser. You type in what you want to log, and the app handles categorising it, analysing it, and showing you insights. A terminal interface is also available for those who prefer it.

## Categories

The app starts with three health categories:

- **Food** -- what you eat and drink
- **Sleep** -- sleep quality, duration, and patterns
- **Activity** -- exercise, movement, and physical activity

More categories can be added over time without any disruption to existing data.

## What's Planned

The app is being built in stages:

1. **Project setup** -- core structure and configuration
2. **Storage** -- database for entries, analyses, and pattern matching
3. **AI connections** -- support for OpenAI, Anthropic, and local AI models
4. **Analysis engine** -- per-category analysis and cross-category insight detection
5. **Web interface** -- browser-based input and analysis display
6. **Terminal interface** -- command-line alternative
7. **Summaries** -- automatic weekly and monthly trend reports
