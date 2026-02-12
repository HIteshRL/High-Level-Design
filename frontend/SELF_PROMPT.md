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
