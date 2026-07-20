package p

// HUMAN: The provider requires ordered retries.
// Reordering these calls breaks failover.
func first() {}

// Retry behavior is tracked in ENG-482.
func second() {}

/*
 * Preserve this behavior.
 * @owned
 */
func third() {}
