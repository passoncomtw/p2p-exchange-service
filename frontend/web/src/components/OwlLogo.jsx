const OwlLogo = ({ size = 48 }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
    focusable="false"
  >
    <rect width="48" height="48" rx="8" fill="#FFC107" />
    <circle cx="17" cy="22" r="7" fill="white" />
    <circle cx="31" cy="22" r="7" fill="white" />
    <circle cx="17" cy="22" r="4" fill="#333" />
    <circle cx="31" cy="22" r="4" fill="#333" />
    <circle cx="18" cy="21" r="1.5" fill="white" />
    <circle cx="32" cy="21" r="1.5" fill="white" />
    <path d="M20 30 Q24 34 28 30" stroke="#333" strokeWidth="1.5" strokeLinecap="round" fill="none" />
    <polygon points="24,12 21,17 27,17" fill="#FF8F00" />
  </svg>
)

export default OwlLogo
