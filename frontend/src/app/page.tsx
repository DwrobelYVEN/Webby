export default function HomePage() {
  return (
    <div>
      <h1>Verified volunteer hours, from first sign-up to final record.</h1>
      <p>
        YVEN connects students, schools, and organizations around a single
        source of truth for community service hours — logged, verified,
        and exportable as an official VSR.
      </p>
      <div style={{ display: 'flex', gap: 12, marginTop: 24 }}>
        <a className="btn" href="/volunteer/signup">Volunteer Sign Up</a>
        <a className="btn" href="/events" style={{ background: '#444' }}>Find Events</a>
      </div>

      <section style={{ marginTop: 48 }}>
        <h2>How it works</h2>
        <div className="card"><strong>1. Register</strong> — create a volunteer profile with your skills and availability.</div>
        <div className="card"><strong>2. Log service</strong> — submit hours tied to a specific event and supervisor.</div>
        <div className="card"><strong>3. Get verified</strong> — your assigned supervisor confirms attendance and hours.</div>
        <div className="card"><strong>4. Export your VSR</strong> — download a locked, audit-backed record of verified hours.</div>
      </section>
    </div>
  );
}
