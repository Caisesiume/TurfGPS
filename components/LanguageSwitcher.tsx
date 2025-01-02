'use client'

import { useLocale } from 'next-intl'
import { useRouter } from 'next/navigation'

export default function LanguageSwitcher() {
  const locale = useLocale()
  const router = useRouter()

  const handleChange = (newLocale: string) => {
    const newPath = window.location.pathname.replace(`/${locale}`, `/${newLocale}`)
    router.push(newPath)
  }

  return (
    <select
      value={locale}
      onChange={(e) => handleChange(e.target.value)}
      className="bg-white text-gray-900 p-2 rounded border border-gray-300"
    >
      <option value="en">English</option>
      <option value="sv">Svenska</option>
    </select>
  )
}