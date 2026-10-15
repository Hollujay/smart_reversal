import React from 'react';
import { AppProvider, useApp } from './context/AppContext';
import { AppShell } from './shell/AppShell';
import { DashboardView } from './features/DashboardView';
import { TransactionsView } from './features/TransactionsView';
import { ReconciliationView } from './features/ReconciliationView';
import { ExceptionsView } from './features/ExceptionsView';
import { InvestigationsView } from './features/InvestigationsView';
import { ReversalsView } from './features/ReversalsView';
import { AuditView } from './features/AuditView';
import { StorybookView } from './features/StorybookView';
import { AgentIntakeView } from './features/AgentIntakeView';
import { ReverseItView } from './features/reverseit/ReverseItView';
import { RulesAndHealthView } from './features/RulesAndHealthView';

const MainViewRouter: React.FC = () => {
  const { activeView } = useApp();

  switch (activeView) {
    case 'dashboard':
      return <DashboardView />;
    case 'agent-intake':
      return <AgentIntakeView />;
    case 'reverseit':
      return <ReverseItView />;
    case 'transactions':
      return <TransactionsView />;
    case 'reconciliation':
      return <ReconciliationView />;
    case 'exceptions':
      return <ExceptionsView />;
    case 'investigations':
      return <InvestigationsView />;
    case 'reversals':
      return <ReversalsView />;
    case 'audit':
      return <AuditView />;
    case 'rules-health':
      return <RulesAndHealthView />;
    case 'storybook':
      return <StorybookView />;
    default:
      return <DashboardView />;
  }
};

export function App() {
  return (
    <AppProvider>
      <AppShell>
        <MainViewRouter />
      </AppShell>
    </AppProvider>
  );
}

export default App;
