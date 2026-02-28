# HealthAnalyzer

HealthAnalyzer is a personal health tracking tool that uses AI to spot patterns and trends across different areas of your daily life.

## How It Works

You provide information about your day in plain language -- what you ate, how you slept, how you're feeling, what activity you did. The app analyses each area separately, then looks across them for connections. For example, it might notice that your sleep quality tends to drop after days with high caffeine intake, that your mood follows patterns linked to your menstrual cycle, or that certain foods consistently trigger digestive discomfort.

You don't need to fill in forms or use a specific format. Just describe your day naturally and the app takes care of the rest.

## What It Tracks

The app covers nine areas of daily health:

- **Sleep** -- total sleep time, REM and deep sleep stages, how often you wake
- **Macro Nutrients** -- what you eat and drink, with automatic tracking of protein, carbs, fat, and calories
- **Cycle Tracker** -- menstrual cycle day and phase, linked to hormonal patterns
- **Digestion** -- how you feel after meals (bloating, heartburn, nausea, etc.)
- **Weight** -- daily or periodic weight tracking
- **Feelings** -- emotional state throughout the day (mood swings, confidence, anxiety, gratitude, and more)
- **Energy** -- energy levels from exhausted to fully energised
- **Mind** -- mental clarity, focus, motivation, stress, and cognitive patterns
- **Activity** -- exercise, movement, and physical activity

More categories can be added over time without any disruption to existing data.

## What It Learns

The app keeps a history of your entries and its own analysis. Each time you submit new information, it actively looks back through your history for relevant patterns -- not just recent days, but older entries that are similar to what you're describing now. If it notices something worth investigating, it can dig deeper into your data on its own.

This means it gets more useful over time as it builds a richer picture of your patterns. Over longer periods, the app produces weekly and monthly summaries so that trends across weeks and months are not lost as daily details accumulate.

## Cross-Category Insights

The real value comes from connections between categories. The app looks for correlations like:

- How your menstrual cycle phase affects mood, energy, and digestion
- Whether poor sleep predicts lower energy and focus the next day
- Which foods are associated with digestive discomfort
- How activity levels relate to sleep quality
- Whether weight trends track with changes in nutrition or activity

These insights are surfaced automatically as part of the daily analysis.

## Your Data, Your Machine

Everything runs locally. Your health data is stored in a single file on your computer. You can choose to use cloud-based AI services (OpenAI or Anthropic) for the analysis, or run a fully local AI model so that nothing leaves your machine. Even with cloud AI, only the text you submit is sent for analysis -- nothing is stored externally.

## How You Use It

The app has a web interface you open in your browser. You type in what you want to log, and the app handles categorising it, analysing it, and showing you insights. A terminal interface is also available for those who prefer it.

## What's Planned

The app is being built in stages:

1. **Project setup** -- core structure and configuration
2. **Storage** -- database for entries, analyses, and pattern matching
3. **AI connections** -- support for OpenAI, Anthropic, and local AI models
4. **Analysis engine** -- per-category analysis, cross-category insight detection, and intelligent data retrieval
5. **Web interface** -- browser-based input and analysis display
6. **Terminal interface** -- command-line alternative
7. **Summaries** -- automatic weekly and monthly trend reports
8. **Meal photos** -- take a photo of your meal and the app analyses it for nutritional content
9. **Device import** -- pull in sleep and activity data from fitness trackers and health apps
