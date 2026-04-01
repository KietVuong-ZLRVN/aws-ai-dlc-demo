import { Routes, Route, Navigate } from 'react-router-dom'
import PreferenceInputScreen from './screens/PreferenceInputScreen'
import ConfirmationScreen from './screens/ConfirmationScreen'
import ComboSuggestionScreen from './screens/ComboSuggestionScreen'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/style/preferences" replace />} />
      <Route path="/style/preferences" element={<PreferenceInputScreen />} />
      <Route path="/style/confirm" element={<ConfirmationScreen />} />
      <Route path="/style/combos" element={<ComboSuggestionScreen />} />
    </Routes>
  )
}
