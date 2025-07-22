document.addEventListener('DOMContentLoaded', function () {
  const form = document.getElementById('backtest-form');
  const resultDiv = document.getElementById('result');

  form.addEventListener('submit', async function (e) {
    e.preventDefault();
    resultDiv.innerHTML = '<span>Loading...</span>';

    // Get form values
    const amount = document.getElementById('amount').value.trim();
    const ticker = document.getElementById('ticker').value.trim();
    const buyDate = document.getElementById('buyDate').value;
    const sellDate = document.getElementById('sellDate').value;
    const type = document.getElementById('type').value;
    const drip = document.getElementById('drip').checked;

    // Build the URL
    let url = '';
    if (sellDate && drip) {
      url = `/${encodeURIComponent(amount)}/of/${encodeURIComponent(ticker)}/on/${buyDate}/and-sold-on/${sellDate}/with-drip?type=${type}`;
    } else if (sellDate) {
      url = `/${encodeURIComponent(amount)}/of/${encodeURIComponent(ticker)}/on/${buyDate}/and-sold-on/${sellDate}?type=${type}`;
    } else {
      url = `/${encodeURIComponent(amount)}/of/${encodeURIComponent(ticker)}/on/${buyDate}?type=${type}`;
    }

    try {
      const response = await fetch(url);
      const data = await response.json();
      if (!response.ok) {
        resultDiv.innerHTML = `<div class="error">❌ <b>Error:</b> ${data.error || 'Unknown error'}<br><small>${data.details || ''}</small></div>`;
        return;
      }
      resultDiv.innerHTML = renderResult(data);
    } catch (err) {
      resultDiv.innerHTML = `<div class="error">❌ <b>Network error:</b> ${err.message}</div>`;
    }
  });
});

function renderResult(data) {
  // Pretty print the result, highlight key fields
  let html = '<h2>Result</h2><ul>';
  for (const [key, value] of Object.entries(data)) {
    if (Array.isArray(value)) {
      html += `<li><b>${key}:</b> <pre>${JSON.stringify(value, null, 2)}</pre></li>`;
    } else {
      html += `<li><b>${key}:</b> ${value}</li>`;
    }
  }
  html += '</ul>';
  return html;
} 