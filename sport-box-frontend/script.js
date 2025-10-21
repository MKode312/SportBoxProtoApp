// API Configuration
const API_BASE_URL = 'http://localhost:8082/api';

// Utility Functions
function getAuthCookie() {
    return document.cookie.split('; ').find(row => row.startsWith('auth_token='))?.split('=')[1];
}

function showLoading() {
    const overlay = document.createElement('div');
    overlay.className = 'spinner-overlay';
    overlay.innerHTML = `
        <div class="spinner-border text-light" role="status">
            <span class="visually-hidden">Loading...</span>
        </div>
    `;
    document.body.appendChild(overlay);
}

function hideLoading() {
    const overlay = document.querySelector('.spinner-overlay');
    if (overlay) {
        overlay.remove();
    }
}

async function apiRequest(endpoint, options = {}, showLoadingIndicator = true) {
    if (showLoadingIndicator) {
        showLoading();
    }

    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };

    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            ...options,
            headers,
            credentials: 'include' // Important for handling cookies
        });

        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || data.message || 'Something went wrong');
        }

        return data;
    } catch (error) {
        console.error('API Request Error:', error);
        throw new Error(error.message || 'Network error occurred');
    } finally {
        if (showLoadingIndicator) {
            hideLoading();
        }
    }
}

// Authentication Functions
async function login(event) {
    event.preventDefault();
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;

    try {
        const response = await apiRequest('/auth/login', {
            method: 'POST',
            body: JSON.stringify({
                email,
                password,
                appID: 1 // Required by the API
            })
        });

        if (response.status === "OK") {
            document.getElementById('userEmail').textContent = email;
            updateAuthUI(true);
            loadMainContent();
        }
    } catch (error) {
        showError('login', error.message);
    }
}

async function register(event) {
    event.preventDefault();
    const email = document.getElementById('registerEmail').value;
    const password = document.getElementById('registerPassword').value;

    try {
        const response = await apiRequest('/auth/register', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });

        if (response.status === "OK") {
            showSuccess('register', 'Registration successful! Please login.');
            showLoginForm();
        }
    } catch (error) {
        showError('register', error.message);
    }
}

function logout() {
    document.cookie = 'auth_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
    updateAuthUI(false);
    location.reload();
}

// UI Update Functions
function updateAuthUI(isAuthenticated) {
    const authButtons = document.getElementById('authButtons');
    const userInfo = document.getElementById('userInfo');
    const authForms = document.getElementById('authForms');
    const mainContent = document.getElementById('mainContent');

    if (isAuthenticated) {
        authButtons.classList.add('d-none');
        userInfo.classList.remove('d-none');
        authForms.classList.add('d-none');
        mainContent.classList.remove('d-none');
    } else {
        authButtons.classList.remove('d-none');
        userInfo.classList.add('d-none');
        authForms.classList.remove('d-none');
        mainContent.classList.add('d-none');
    }
}

function showLoginForm() {
    document.getElementById('loginForm').classList.remove('d-none');
    document.getElementById('registerForm').classList.add('d-none');
}

function showRegisterForm() {
    document.getElementById('loginForm').classList.add('d-none');
    document.getElementById('registerForm').classList.remove('d-none');
}

function showError(formType, message) {
    let container;
    if (formType === 'main') {
        // Create or get the main error container
        container = document.querySelector('.main-error-container');
        if (!container) {
            container = document.createElement('div');
            container.className = 'main-error-container alert alert-danger alert-dismissible fade show mt-3';
            document.querySelector('.container').prepend(container);
        }
    } else {
        const form = document.getElementById(`${formType}Form`);
        if (!form) return;
        
        // Create or get the form error container
        container = form.querySelector('.error-message');
        if (!container) {
            container = document.createElement('div');
            container.className = 'error-message alert alert-danger mt-3';
            form.querySelector('form').appendChild(container);
        }
    }

    container.innerHTML = `
        ${message}
        ${formType === 'main' ? '<button type="button" class="btn-close" data-bs-dismiss="alert"></button>' : ''}
    `;

    // Auto-hide after 5 seconds if it's not the main error
    if (formType !== 'main') {
        setTimeout(() => {
            container.remove();
        }, 5000);
    }
}

function showSuccess(formType, message) {
    let container;
    if (formType === 'main') {
        // Create or get the main success container
        container = document.querySelector('.main-success-container');
        if (!container) {
            container = document.createElement('div');
            container.className = 'main-success-container alert alert-success alert-dismissible fade show mt-3';
            document.querySelector('.container').prepend(container);
        }
    } else {
        const form = document.getElementById(`${formType}Form`);
        if (!form) return;
        
        // Create or get the form success container
        container = form.querySelector('.success-message');
        if (!container) {
            container = document.createElement('div');
            container.className = 'success-message alert alert-success mt-3';
            form.querySelector('form').appendChild(container);
        }
    }

    container.innerHTML = `
        ${message}
        ${formType === 'main' ? '<button type="button" class="btn-close" data-bs-dismiss="alert"></button>' : ''}
    `;

    // Auto-hide after 5 seconds if it's not the main success message
    if (formType !== 'main') {
        setTimeout(() => {
            container.remove();
        }, 5000);
    }
}

// Payment Functions
async function addCard(event) {
    event.preventDefault();
    const email = document.getElementById('userEmail').textContent;
    const cardNumber = parseInt(document.getElementById('cardNumber').value);
    const cvc = parseInt(document.getElementById('cvc').value);
    const phoneNumber = parseInt(document.getElementById('phoneNumber').value);

    try {
        const response = await apiRequest('/payments/add-card', {
            method: 'POST',
            body: JSON.stringify({
                email,
                cardNumber,
                cvc,
                phoneNumber
            })
        });

        if (response.success) {
            showSuccess('payment', 'Card added successfully!');
            await loadPaymentMethods(); // Refresh payment methods display
            paymentModal.hide();
        }
    } catch (error) {
        showError('payment', error.message);
    }
}

// Booking Functions
async function createBooking(event) {
    event.preventDefault();
    const email = document.getElementById('userEmail').textContent;
    const boxName = document.getElementById('boxName').value;
    const peopleAmount = parseInt(document.getElementById('peopleAmount').value);
    const timeStart = document.getElementById('timeStart').value;
    const timeHrs = parseInt(document.getElementById('timeHrs').value);
    const timeMins = parseInt(document.getElementById('timeMins').value);

    try {
        const response = await apiRequest('/book', {
            method: 'POST',
            body: JSON.stringify({
                email,
                boxName,
                peopleAmount,
                timeStart,
                timeHrs,
                timeMins
            })
        });

        if (response.success) {
            showSuccess('booking', `Booking successful! Reservation ID: ${response.resID}`);
            updateBalance(response.balance);
            bookingModal.hide();
            loadBoxes();
            loadBookings();
        }
    } catch (error) {
        showError('booking', error.message);
    }
}

// Payment Methods Functions
async function loadPaymentMethods() {
    const email = document.getElementById('userEmail').textContent;
    
    try {
        const response = await apiRequest('/payments/cards?email=' + encodeURIComponent(email), {
            method: 'GET'
        });

        const cardsList = document.getElementById('cardsList');
        if (!cardsList) return;

        if (response.cardNumber && response.phoneNumber) {
            // Mask card number to show only last 4 digits
            const maskedCard = '**** **** **** ' + response.cardNumber.toString().slice(-4);
            // Mask phone number to show only last 4 digits
            const maskedPhone = '*******' + response.phoneNumber.toString().slice(-4);
            
            cardsList.innerHTML = `
                <div class="col-md-6">
                    <div class="card mb-3">
                        <div class="card-body">
                            <h5 class="card-title">Card Details</h5>
                            <p class="card-text">
                                <strong>Card Number:</strong> ${maskedCard}<br>
                                <strong>Phone Number:</strong> ${maskedPhone}
                            </p>
                        </div>
                    </div>
                </div>
            `;
        } else {
            cardsList.innerHTML = `
                <div class="col-12">
                    <div class="alert alert-info">
                        No payment methods found. Add a card to make bookings.
                    </div>
                </div>
            `;
        }
    } catch (error) {
        const cardsList = document.getElementById('cardsList');
        if (cardsList) {
            cardsList.innerHTML = `
                <div class="col-12">
                    <div class="alert alert-info">
                        No payment methods found. Add a card to make bookings.
                    </div>
                </div>
            `;
        }
        // Don't show error for no cards, it's a normal state
        if (error.message !== 'Card not found') {
            console.error('Error loading payment methods:', error);
            showError('payment', 'Failed to load payment methods: ' + error.message);
        }
    }
}

// Modal Functions
let bookingModal, paymentModal;

function openBookingModal(boxName) {
    document.getElementById('boxName').value = boxName;
    populateTimeSlots();
    bookingModal.show();
}

function closeBookingModal() {
    bookingModal.hide();
}

function openPaymentModal() {
    paymentModal.show();
}

function closePaymentModal() {
    paymentModal.hide();
}

function populateTimeSlots() {
    const timeSelect = document.getElementById('timeStart');
    timeSelect.innerHTML = '<option value="">Select time...</option>';
    
    for (let hour = 8; hour < 22; hour++) {
        for (let min = 0; min < 60; min += 30) {
            const timeValue = `${hour.toString().padStart(2, '0')}:${min.toString().padStart(2, '0')}`;
            timeSelect.add(new Option(timeValue, timeValue));
        }
    }
}

async function loadBoxes() {
    try {
        const response = await apiRequest('/boxes');
        const container = document.getElementById('boxesContainer');
        container.innerHTML = '';

        response.boxes.forEach(box => {
            container.appendChild(createBoxCard(box));
        });
    } catch (error) {
        console.error('Error loading boxes:', error);
        showError('main', 'Failed to load boxes');
    }
}

function createBoxCard(box) {
    const card = document.createElement('div');
    card.className = 'col-md-6 col-lg-4 mb-4';
    card.innerHTML = `
        <div class="card box-card" onclick="openBookingModal('${box.name}')">
            <div class="card-body">
                <h5 class="card-title">${box.name}</h5>
                <div class="box-price">$${box.pricePerHour}/hour</div>
                <div class="box-status ${box.available ? 'status-available' : 'status-booked'}">
                    ${box.available ? 'Available' : 'Booked'}
                </div>
            </div>
        </div>
    `;
    return card;
}

async function loadBookings() {
    try {
        const response = await apiRequest('/bookings');
        const container = document.getElementById('bookingsList');
        container.innerHTML = '';

        response.bookings.forEach(booking => {
            container.appendChild(createBookingItem(booking));
        });
    } catch (error) {
        console.error('Error loading bookings:', error);
        showError('main', 'Failed to load bookings');
    }
}

function createBookingItem(booking) {
    const item = document.createElement('div');
    item.className = 'booking-item';
    item.innerHTML = `
        <div>
            <strong>${booking.boxName}</strong>
            <p>Time: ${booking.timeStart}</p>
            <p>Duration: ${booking.timeHrs}h ${booking.timeMins}m</p>
            <p>People: ${booking.peopleAmount}</p>
            <div class="booking-actions">
                <button class="btn btn-sm btn-danger" onclick="cancelBooking(${booking.id})">
                    Cancel
                </button>
            </div>
        </div>
    `;
    return item;
}

async function cancelBooking(bookingId) {
    if (!confirm('Are you sure you want to cancel this booking?')) {
        return;
    }

    try {
        await apiRequest(`/bookings/${bookingId}`, {
            method: 'DELETE'
        });

        loadBoxes();
        loadBookings();
    } catch (error) {
        console.error('Error canceling booking:', error);
        showError('main', 'Failed to cancel booking');
    }
}

// Initialization
document.addEventListener('DOMContentLoaded', async () => {
    bookingModal = new bootstrap.Modal(document.getElementById('bookingModal'));
    paymentModal = new bootstrap.Modal(document.getElementById('paymentModal'));
    
    // Check if token is valid by making a request to a protected endpoint
    try {
        const authCookie = getAuthCookie();
        if (authCookie) {
            // Try to load user data with the token
            await apiRequest('/payments/cards', {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${authCookie}`
                }
            }, false);
            
            updateAuthUI(true);
            loadMainContent();
            loadPaymentMethods();
        }
    } catch (error) {
        console.log('Session expired or invalid', error);
        // If the request fails, clear the cookie and show login
        document.cookie = 'auth_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
        updateAuthUI(false);
    }
});

async function loadMainContent() {
    try {
        // Load boxes and bookings separately to handle errors individually
        try {
            await loadBoxes();
        } catch (error) {
            console.error('Error loading boxes:', error);
            showError('main', 'Failed to load available boxes. Please try again later.');
        }

        try {
            await loadBookings();
        } catch (error) {
            console.error('Error loading bookings:', error);
            // Show "No bookings found" instead of error if appropriate
            const bookingsList = document.getElementById('bookingsList');
            if (bookingsList) {
                bookingsList.innerHTML = `
                    <div class="alert alert-info">
                        No active bookings found.
                    </div>
                `;
            }
        }
    } catch (error) {
        console.error('Error in loadMainContent:', error);
    }
}
