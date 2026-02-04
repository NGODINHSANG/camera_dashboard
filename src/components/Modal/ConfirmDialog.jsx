function ConfirmDialog({ message, itemName, onConfirm, onCancel, confirmText = 'Xác nhận', cancelText = 'Hủy', type = 'danger' }) {
    return (
        <div>
            <div className="confirm-icon">
                <svg viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
                </svg>
            </div>
            <p className="confirm-message">
                {message}
                {itemName && (
                    <>
                        <br />
                        <strong>{itemName}</strong>
                    </>
                )}
            </p>
            <div className="modal-actions" style={{ justifyContent: 'center' }}>
                <button className="btn btn-secondary" onClick={onCancel}>
                    {cancelText}
                </button>
                <button
                    className={`btn ${type === 'danger' ? 'btn-danger' : 'btn-primary'}`}
                    onClick={onConfirm}
                >
                    {confirmText}
                </button>
            </div>
        </div>
    )
}

export default ConfirmDialog
