import { useState, useEffect } from 'react'

/**
 * Custom hook để lưu state vào localStorage
 * @param {string} key - Key trong localStorage
 * @param {any} initialValue - Giá trị khởi tạo nếu chưa có trong localStorage
 */
function useLocalStorage(key, initialValue) {
    // Lấy giá trị từ localStorage hoặc dùng initialValue
    const [storedValue, setStoredValue] = useState(() => {
        try {
            const item = window.localStorage.getItem(key)
            return item ? JSON.parse(item) : initialValue
        } catch (error) {
            console.error(`Error reading localStorage key "${key}":`, error)
            return initialValue
        }
    })

    // Lưu vào localStorage mỗi khi giá trị thay đổi
    useEffect(() => {
        try {
            window.localStorage.setItem(key, JSON.stringify(storedValue))
        } catch (error) {
            console.error(`Error setting localStorage key "${key}":`, error)
        }
    }, [key, storedValue])

    return [storedValue, setStoredValue]
}

export default useLocalStorage
