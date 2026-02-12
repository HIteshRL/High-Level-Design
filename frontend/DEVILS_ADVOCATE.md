# Devil's Advocate Review (Frontend)

## 1) Could this architecture be over-engineered?
- **Risk:** Adding too many abstractions early slows feature delivery.
- **Decision:** Keep MVP with local state and typed API module; defer heavy global state libraries.

## 2) Could token storage in localStorage be abused?
- **Risk:** XSS could expose token.
- **Decision:** Accept for MVP/local tool. For production, move to httpOnly cookies + CSRF strategy.

## 3) Could UI claim features backend doesn't support yet?
- **Risk:** Mismatch creates user confusion.
- **Decision:** Avoid streaming UI and advanced controls until backend guarantees behavior.

## 4) Could conversation switching be misleading without history fetch API?
- **Risk:** Users expect historic messages when selecting a conversation.
- **Decision:** Clear local behavior for now; next iteration should add `GET /conversations/:id/messages`.

## 5) Could error handling hide useful debugging info?
- **Risk:** Generic errors harm debugging.
- **Decision:** Show friendly messages to users; keep structured details in browser dev tools/network.

## 6) Could design system drift happen quickly?
- **Risk:** Ad-hoc styles break consistency.
- **Decision:** Force new UI through shared primitives and tokenized styles.

## Convergence Outcome
- Current implementation balances speed, maintainability, and accuracy for MVP.
- Highest-value next steps: message history endpoint integration, component tests, accessibility pass.
