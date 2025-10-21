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
    
    // Get values from form
    const email = document.getElementById('userEmail').textContent;
    const cardNumberInput = document.getElementById('cardNumber').value.replace(/\D/g, ''); // Remove non-digits
    const cvcInput = document.getElementById('cvc').value.replace(/\D/g, ''); // Remove non-digits
    const phoneNumberInput = document.getElementById('phoneNumber').value.replace(/\D/g, ''); // Remove non-digits

    // Validate inputs
    if (cardNumberInput.length !== 16) {
        showError('payment', 'Card number must be 16 digits');
        return;
    }
    if (cvcInput.length !== 3) {
        showError('payment', 'CVC must be 3 digits');
        return;
    }
    if (phoneNumberInput.length !== 10) {
        showError('payment', 'Phone number must be 10 digits');
        return;
    }

    // Convert to numbers
    const cardNumber = Number(cardNumberInput);
    const cvc = Number(cvcInput);
    const phoneNumber = Number(phoneNumberInput);

    try {
        const response = await apiRequest('/payments/add-card', {
            method: 'POST',
            body: JSON.stringify({
                email,
                cardNumber: cardNumber,
                cvc: cvc,
                phoneNumber: phoneNumber
            })
        });

        if (response.error) {
            throw new Error(response.error);
        }

        showSuccess('main', 'Card added successfully!');
        document.getElementById('paymentForm').reset();
        await loadPaymentMethods(); // Refresh payment methods display
        paymentModal.hide();
    } catch (error) {
        console.error('Add card error:', error);
        showError('main', 'Failed to add card: ' + error.message);
    }
}

// Booking Functions
async function createBooking(event) {
    event.preventDefault();
    
    // Validate that user has added a payment method first
    try {
        const cardResponse = await apiRequest('/payments/cards?email=' + encodeURIComponent(document.getElementById('userEmail').textContent), {
            method: 'GET'
        });
        
        if (!cardResponse || (Array.isArray(cardResponse) && cardResponse.length === 0)) {
            showError('booking', 'Please add a payment method before booking');
            return;
        }
    } catch (error) {
        showError('booking', 'Please add a payment method before booking');
        return;
    }

    // Validate form inputs
    const email = document.getElementById('userEmail').textContent;
    const boxName = document.getElementById('boxName').value;
    const peopleAmount = parseInt(document.getElementById('peopleAmount').value);
    const timeStart = document.getElementById('timeStart').value;
    const timeHrs = parseInt(document.getElementById('timeHrs').value);
    const timeMins = parseInt(document.getElementById('timeMins').value);

    // Input validation
    if (!timeStart) {
        showError('booking', 'Please select a start time');
        return;
    }

    if (isNaN(peopleAmount) || peopleAmount < 1) {
        showError('booking', 'Please enter a valid number of people');
        return;
    }

    if (isNaN(timeHrs) || timeHrs < 0 || (timeHrs === 0 && timeMins === 0)) {
        showError('booking', 'Please enter a valid duration');
        return;
    }

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

        if (response.error) {
            throw new Error(response.error);
        }

        // Show success message and update UI
        showSuccess('main', `Booking successful! Reservation ID: ${response.resID || 'N/A'}`);
        if (response.balance !== undefined) {
            updateBalance(response.balance);
        }
        bookingModal.hide();
        
        // Refresh the boxes and bookings lists
        await Promise.all([
            loadBoxes(),
            loadBookings()
        ]);
        
        // Clear form
        document.getElementById('bookingForm').reset();
        
    } catch (error) {
        console.error('Booking error:', error);
        const errorMessage = error.message || 'Failed to create booking. Please try again.';
        showError('main', `Booking Error: ${errorMessage}`);
        // Keep the modal open so user can see the error
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

        console.log('Cards response:', response); // Debug log

        // Handle different response formats
        let cards = [];
        if (response) {
            if (Array.isArray(response)) {
                cards = response;
            } else if (typeof response === 'object') {
                // If it's a single card object
                cards = [response];
            }
        }

        if (cards && cards.length > 0) {
            cardsList.innerHTML = cards.map(card => {
                try {
                    // Handle both string and number formats
                    const cardNumber = typeof card.cardNumber === 'int' ? 
                        card.cardNumber.replace(/[^\d]/g, '') : // Remove non-digits if string
                        (card.cardNumber ? card.cardNumber.toString() : '');
                    
                    const phoneNumber = typeof card.phoneNumber === 'int' ?
                        card.phoneNumber.replace(/[^\d]/g, '') : // Remove non-digits if string
                        (card.phoneNumber ? card.phoneNumber.toString() : '');

                    // Only show the card if we have valid numbers
                    if (!cardNumber) {
                        console.warn('Invalid card number received:', card.cardNumber);
                        return '';
                    }

                    // Format the display
                    const maskedCard = cardNumber.length >= 4 ? 
                        '**** **** **** ' + cardNumber.slice(-4) : 
                        '**** **** **** ****';
                    
                    const maskedPhone = phoneNumber.length >= 4 ?
                        '*******' + phoneNumber.slice(-4) :
                        '***********';

                    return `
                        <div class="col-md-6">
                            <div class="card mb-3">
                                <div class="card-body">
                                    <h5 class="card-title">Payment Card</h5>
                                    <p class="card-text">
                                        <strong>Card Number:</strong> ${maskedCard}<br>
                                        <strong>Phone Number:</strong> ${maskedPhone}
                                    </p>
                                </div>
                            </div>
                        </div>
                    `;
                } catch (err) {
                    console.error('Error processing card:', err);
                    return '';
                }
            }).filter(html => html !== '').join('');

            if (cardsList.innerHTML === '') {
                cardsList.innerHTML = `
                    <div class="col-12">
                        <div class="alert alert-warning">
                            No valid cards found. Please add a new card.
                        </div>
                    </div>
                `;
            }
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
        console.error('Error loading payment methods:', error);
        const cardsList = document.getElementById('cardsList');
        if (cardsList) {
            cardsList.innerHTML = `
                <div class="col-12">
                    <div class="alert alert-danger">
                        <strong>Error loading cards:</strong> ${error.message}
                    </div>
                </div>
            `;
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
            <div class="box-price">₽${box.pricePerHour}/hour</div>
                <img src="assets/sportBoxImage.jpg" alt="${box.name}" class="box-image">
                <h5 class="card-title">${box.name}</h5>
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
