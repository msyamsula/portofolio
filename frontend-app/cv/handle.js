// EmailJS Configuration
// Get these from https://www.emailjs.com/
const EMAILJS_PUBLIC_KEY = '9BW53Ryn0F6h86V5F'; // Replace with your public key
const EMAILJS_SERVICE_ID = 'service_kbnquew'; // Replace with your service ID
const EMAILJS_TEMPLATE_ID = 'template_xrky1jd'; // Replace with your template ID

function sayHello() {
  console.log("hello world, from script");
  return "hello world, from script";
}

async function submitInquiry(sender, email, message, subject) {
  // Validate inputs
  if (!sender || !email || !message || !subject) {
    console.error("Missing required fields");
    return Promise.reject(new Error("All fields are required"));
  }

  // Validate email format
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) {
    console.error("Invalid email format");
    return Promise.reject(new Error("Invalid email format"));
  }

  try {
    // Initialize EmailJS if not already initialized
    if (typeof emailjs !== 'undefined' && !emailjs._initialized) {
      emailjs.init(EMAILJS_PUBLIC_KEY);
      emailjs._initialized = true;
    }

    // Send email using EmailJS
    const response = await emailjs.send(
      EMAILJS_SERVICE_ID,
      EMAILJS_TEMPLATE_ID,
      {
        name: sender,
        email: email,
        title: subject,
        message: message,
      }
    );

    console.log("Email sent successfully!", response);
    return {
      success: true,
      message: "Message sent successfully!"
    };

  } catch (error) {
    console.error("Error sending email:", error);
    return Promise.reject(new Error("Failed to send message: " + (error.text || error.message)));
  }
}
