# LOKIX - A go based web crawler [WIP]
NOTE: _work in progress_

A project to learn more about web crawler internals. Topics to cover:
- What is a web-crawler? How does it work?
- Architecture of a monolithic web-crawler
- Embedding search capabilities using inverted-indexes
- Building the web-crawler from scratch -> the URL queue, the crawler, the HTML content parser, the index and inverted index maker? (idk yet) tests? (how? TvT)
- Tech Stack
- Design Decisions and Tradeoffs
- Pitfalls/Shortcomings of the Implementation + Future Scope

### Setup
- Docker -> `docker compose up --build` to rebuild the image if there are any changes, otherwise `docker compose up` will run the latest image

- Testing (via Docker) -> `docker compose -f compose.test.yml up --build` to run the tests in the container (NOTE, this is a separate stage and if I run the main compose file it will still compile even if tests fail)

NOTE TO SELF: If you want the build to fail if tests fail, you should run the tests within the build-stage before the binary is built, or ensure your CI pipeline specifically targets the run-test-stage first.