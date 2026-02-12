# Self-Prompt for High-Accuracy Frontend Delivery

Use this prompt before changing UI architecture or behavior:

1. **Contract Accuracy**
   - Are API payload keys exactly aligned with backend JSON contracts?
   - Are all nullable fields handled safely in UI rendering?

2. **State Correctness**
   - What state is source-of-truth for auth token, user, conversations, and current messages?
   - Is stale state possible after logout/login or conversation switching?

3. **Failure Modes**
   - What happens on 401/403/404/429/500 from backend?
   - Are errors surfaced to users with actionable feedback?

4. **UX Quality**
   - Are loading, empty, and success states all explicitly visible?
   - Is the primary action always obvious?

5. **Security and Safety**
   - Is sensitive data avoided in logs and visible UI?
   - Are tokens stored/retrieved consistently and removed on logout?

6. **Maintainability**
   - Are components small, typed, and reusable?
   - Is business logic separated from presentational UI?

7. **Performance Baseline**
   - Are expensive re-renders minimized for message lists?
   - Are unnecessary fetches avoided?

8. **Release Readiness**
   - Does the app build cleanly with strict TypeScript?
   - Does dev setup work from a clean checkout with minimal steps?

9. **Theme and Visual System**
   - Are light and dark themes both readable with proper contrast?
   - Does theme preference persist and apply before first meaningful paint?

10. **Advanced Controls Safety**
   - Do model/temperature/max-token controls map exactly to backend request keys?
   - Are control defaults sensible for local inference performance?

11. **Message Rendering Fidelity**
   - Does markdown render safely and legibly for headings/lists/code/links/tables?
   - Are long code blocks and tables readable on small screens?

12. **Input Ergonomics**
   - Does Enter submit only when intended (and not during IME composition)?
   - Is Shift+Enter newline behavior preserved consistently?

13. **History Persistence Contract**
   - Does conversation selection fetch persisted messages from backend every time?
   - Is the active conversation hydrated on initial app load with deterministic behavior?

14. **Sidebar Interaction Safety**
   - Are resize and collapse controls keyboard/mouse safe and bounded by min/max widths?
   - Are width/collapse preferences persisted without breaking responsive mobile layout?

15. **Code Rendering Quality**
   - Are fenced code blocks syntax-highlighted and horizontally scrollable?
   - Is inline code still rendered with compact, readable styling?

16. **Cache Semantics Awareness**
   - Is the backend cache key based on full context (history + prompt + model params), not prompt alone?
   - Could any stale or cross-context answer be served for semantically different histories?

17. **API Evolution Robustness**
   - Are new message-history endpoint failures handled distinctly (`403`/`404`/`5xx`)?
   - Is fallback behavior explicit (empty state vs silent reset)?
