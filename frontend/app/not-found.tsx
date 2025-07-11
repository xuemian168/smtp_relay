export default function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-gray-900 dark:text-gray-100">
          404
        </h1>
        <p className="mt-2 text-lg text-gray-600 dark:text-gray-400">
          页面未找到 / Page Not Found
        </p>
        <a
          href="/"
          className="mt-4 inline-block text-blue-600 hover:text-blue-500"
        >
          返回首页 / Back to Home
        </a>
      </div>
    </div>
  )
} 