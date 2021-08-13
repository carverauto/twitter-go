// streaming-api/app.js

const APP_KEY = 'PUSHER_APP_KEY';
const APP_CLUSTER = 'PUSHER_APP_CLUSTER';

var ctx = document.getElementById('myChart').getContext('2d');
var myChart = new Chart(ctx, {
    type: 'bar',
    data: {
        labels: [],
        datasets: [
            {
                label: '# of Tweets',
                data: [],
                backgroundColor: [
                    'rgba(255, 99, 132, 0.2)',
                    'rgba(54, 162, 235, 0.2)',
                    'rgba(255, 159, 64, 0.2)',
                ],
                borderWidth: 1,
            },
        ],
    },
    options: {
        scales: {
            yAxes: [
                {
                    ticks: {
                        beginAtZero: true,
                    },
                },
            ],
        },
    },
});

function updateChart(data) {
    let iterationCount = 0;

    for (const key in data) {
        if (!myChart.data.labels.includes(key)) {
            myChart.data.labels.push(key);
        }

        myChart.data.datasets.forEach(dataset => {
            dataset.data[iterationCount] = data[key];
        });

        iterationCount++;

        myChart.update();
    }
}

axios
    .get('http://localhost:1500/polls', {})
    .then(res => {
        updateChart(res.data);
    })
    .catch(err => {
        console.log('Could not retrieve information from the backend');
        console.error(err);
    });

const pusher = new Pusher(APP_KEY, {
    cluster: APP_CLUSTER,
});

const channel = pusher.subscribe('twitter-votes');

channel.bind('options', data => {
    updateChart(data);
});
