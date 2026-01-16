# 🌤️ Vremenska Prognoza - Network Version

## 📁 What's Included
- `weather-app.exe` - Standalone executable (no installation required)
- Embedded HTML/CSS files (all assets are included in the exe)

## 🚀 How to Run on Network PCs

### Method 1: Direct Network Share
1. **Copy the executable** to a shared network folder
2. **From any PC on your network**, navigate to: `\\YourPC\SharedFolder\weather-app.exe`
3. **Double-click** to run

### Method 2: USB Drive Transfer
1. **Copy `weather-app.exe`** to a USB drive
2. **Paste on any PC** and double-click to run

### Method 3: Network Download
1. **Share the executable** via network share or cloud storage
2. **Download and run** on any PC

## 🌐 Network Access

Once running, the weather dashboard will be accessible from **any device on your network** at:

- **Local access**: `http://localhost:8081`
- **Network access**: `http://[PC-IP-ADDRESS]:8081`

### Finding the PC IP Address
The application will display the network IP when started, or:
- **Windows**: Open Command Prompt and type `ipconfig`
- **Look for**: "IPv4 Address" (usually 192.168.x.x or 10.x.x.x)

## 📱 Cross-Device Access

From **any device** on your network (phones, tablets, other computers):
1. **Open web browser**
2. **Go to**: `http://[PC-IP-ADDRESS]:8081`
3. **Enjoy the weather dashboard!**

## 🔧 Requirements

- **Windows PC** (Windows 7 or later)
- **Internet connection** (for weather data)
- **Network access** (for other devices to connect)

## ✅ Features

- **Real weather data** for Croatian cities
- **5-day forecasts** 
- **Beautiful responsive design**
- **Dark/Light mode toggle**
- **No installation required**
- **Works on all devices** via web browser

## 🌍 Supported Cities
- Zagreb 🏛️
- Split 🏖️  
- Dubrovnik ⛱️
- Rijeka 🌊
- Zadar 🐚
- Osijek 🌾

## 🚨 Important Notes

- **Firewall**: Windows may ask for firewall permission - allow access
- **Port 8081**: Must be open on the host PC's firewall
- **One instance**: Only run one copy per PC to avoid port conflicts

## 🛠️ Troubleshooting

### "Port already in use"
- Close any other instances of the weather app
- Change port in the source code if needed

### "Cannot access from other devices"
- Check Windows Firewall settings
- Ensure devices are on the same network
- Verify the IP address is correct

### "No weather data"
- Check internet connection
- Wait a few moments for API data to load

---

**Enjoy your weather dashboard across the entire network! 🌤️**
