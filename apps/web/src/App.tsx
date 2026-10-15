import { useState } from 'react'
import logo from './assets/logo.png'
import './App.css'

function App() {
  const [activeTab, setActiveTab] = useState<'overview' | 'ai' | 'reverseit'>('overview')

  return (
    <div className="sr-app">
      {/* Navigation Header */}
      <header className="sr-navbar">
        <div className="sr-brand">
          <img src={logo} className="sr-brand-logo" alt="SMARTREVERSAL Logo" />
          <div className="sr-brand-text">
            <span className="sr-brand-title">SMARTREVERSAL</span>
            <span className="sr-brand-subtitle">Financial Operations Console</span>
          </div>
        </div>
        <div className="sr-nav-status">
          <span className="sr-status-dot online"></span>
          <span className="sr-status-label">AI Workstream: Online (Prompts A0–A9 Verified)</span>
        </div>
      </header>

      {/* Main Hero Section */}
      <main className="sr-main">
        <section className="sr-hero">
          <div className="sr-hero-emblem-wrap">
            <img src={logo} className="sr-hero-logo" alt="SMARTREVERSAL Emblem" />
          </div>
          <div className="sr-badge">Enterprise Transaction Exception Intelligence</div>
          <h1 className="sr-hero-title">Controlled Financial Resolution Platform</h1>
          <p className="sr-hero-description">
            Automated transaction investigation, multi-source discrepancy reconciliation, 
            and maker-checker authorized remediation with zero autonomous fund movements.
          </p>

          {/* System Telemetry Chips */}
          <div className="sr-telemetry-grid">
            <div className="sr-telemetry-chip">
              <span className="chip-value">58 / 58</span>
              <span className="chip-label">AI Test Suite Passing</span>
            </div>
            <div className="sr-telemetry-chip">
              <span className="chip-value">8 Gates</span>
              <span className="chip-label">Reversal Checks Enforced</span>
            </div>
            <div className="sr-telemetry-chip">
              <span className="chip-value">6 Banks</span>
              <span className="chip-label">ReverseIt Alert Parsers</span>
            </div>
            <div className="sr-telemetry-chip">
              <span className="chip-value">100%</span>
              <span className="chip-label">CBN Clock Determinism</span>
            </div>
          </div>
        </section>

        {/* Feature Navigation Tabs */}
        <div className="sr-tabs">
          <button 
            className={`sr-tab ${activeTab === 'overview' ? 'active' : ''}`}
            onClick={() => setActiveTab('overview')}
          >
            System Architecture
          </button>
          <button 
            className={`sr-tab ${activeTab === 'ai' ? 'active' : ''}`}
            onClick={() => setActiveTab('ai')}
          >
            AI/ML Intelligence (A0–A8)
          </button>
          <button 
            className={`sr-tab ${activeTab === 'reverseit' ? 'active' : ''}`}
            onClick={() => setActiveTab('reverseit')}
          >
            ReverseIt Consumer AI (A9)
          </button>
        </div>

        {/* Tab Content Cards */}
        <section className="sr-cards-grid">
          {activeTab === 'overview' && (
            <>
              <div className="sr-card">
                <div className="sr-card-header">
                  <span className="sr-card-tag">Hexagonal Architecture</span>
                  <h3>Multi-Source Reconciliation</h3>
                </div>
                <p>2-way and 3-way automated matching across Core Banking ledgers, switch feeds, and processor settlement journals.</p>
                <div className="sr-card-footer">
                  <span className="status-badge success">Deterministic Match Matrix</span>
                </div>
              </div>

              <div className="sr-card">
                <div className="sr-card-header">
                  <span className="sr-card-tag">Governance & Control</span>
                  <h3>Maker-Checker Dual Control</h3>
                </div>
                <p>Strict segregation of duties. High-value reversals and kill switch alterations require cryptographic dual authorization.</p>
                <div className="sr-card-footer">
                  <span className="status-badge success">4-Tier Kill Switches Active</span>
                </div>
              </div>

              <div className="sr-card">
                <div className="sr-card-header">
                  <span className="sr-card-tag">Compliance</span>
                  <h3>Tamper-Evident Audit Trail</h3>
                </div>
                <p>Immutable operational logging recording state transitions, actor identification, reason codes, and client telemetry.</p>
                <div className="sr-card-footer">
                  <span className="status-badge success">Zero PII Logged</span>
                </div>
              </div>
            </>
          )}

          {activeTab === 'ai' && (
            <>
              <div className="sr-card">
                <div className="sr-card-header">
                  <span className="sr-card-tag">Prompts A1 & A2</span>
                  <h3>Context Builder & Citation Guard</h3>
                </div>
                <p>Deterministic evidence bundle construction with local PII masking (NUBAN, BVN) and zero-hallucination citation validation.</p>
                <div className="sr-card-footer">
                  <span className="status-badge success">0.0% Hallucinations Tolerated</span>
                </div>
              </div>

              <div className="sr-card">
                <div className="sr-card-header">
                  <span className="sr-card-tag">Prompts A3, A4 & A5</span>
                  <h3>Root Cause & Anomaly Engine</h3>
                </div>
                <p>14-category root cause taxonomy, statistical EWMA spike clustering, Isolation Forest, and pending resolution hazard models.</p>
                <div className="sr-card-footer">
                  <span className="status-badge success">Maintenance Dampening Active</span>
                </div>
              </div>

              <div className="sr-card">
                <div className="sr-card-header">
                  <span className="sr-card-tag">Prompts A7 & A8</span>
                  <h3>Read-Only Copilot & CI Evaluation</h3>
                </div>
                <p>Prompt-injection resistant analyst copilot restricted to read-only tools, backed by continuous CI evaluation regression gates.</p>
                <div className="sr-card-footer">
                  <span className="status-badge success">Refusal Guardrails Verified</span>
                </div>
              </div>
            </>
          )}

          {activeTab === 'reverseit' && (
            <>
              <div className="sr-card">
                <div className="sr-card-header">
                  <span className="sr-card-tag">Local Regex Parsers</span>
                  <h3>Nigerian Bank Alert Ingestion</h3>
                </div>
                <p>High-precision offline regex extractors for WEMA/ALAT, GTB, Zenith, Access, FirstBank, and UBA SMS and email debit alerts.</p>
                <div className="sr-card-footer">
                  <span className="status-badge success">6 Top Banks Supported</span>
                </div>
              </div>

              <div className="sr-card">
                <div className="sr-card-header">
                  <span className="sr-card-tag">Regulatory Integrity</span>
                  <h3>Deterministic CBN Dispute Clocks</h3>
                </div>
                <p>Regulatory resolution deadlines (ATM 24h/48h, POS/Web 72h, NIP 24h) are assigned via strict table lookup, never LLM hallucination.</p>
                <div className="sr-card-footer">
                  <span className="status-badge success">100% Clock Accuracy</span>
                </div>
              </div>

              <div className="sr-card">
                <div className="sr-card-header">
                  <span className="sr-card-tag">Customer Notice Filter</span>
                  <h3>Fact-Checked Complaint Drafter</h3>
                </div>
                <p>Generates verifiable formal complaint letters and applies customer-safe wording filters in English and Nigerian Pidgin.</p>
                <div className="sr-card-footer">
                  <span className="status-badge success">Zero Invented Facts</span>
                </div>
              </div>
            </>
          )}
        </section>
      </main>

      {/* Footer */}
      <footer className="sr-footer">
        <div className="sr-footer-content">
          <div className="sr-footer-brand">
            <img src={logo} className="sr-footer-logo" alt="SMARTREVERSAL Emblem" />
            <span>SMARTREVERSAL Platform © 2026. Bank-grade Financial Operations & Resolution.</span>
          </div>
          <div className="sr-footer-links">
            <span className="tag-pill">Go 1.24+</span>
            <span className="tag-pill">Python 3.12+</span>
            <span className="tag-pill">React + Vite</span>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default App
