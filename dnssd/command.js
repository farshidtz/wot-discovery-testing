#!/usr/bin/env node

const yargs = require('yargs/yargs')
const { hideBin } = require('yargs/helpers')
const MDNSWoTTest = require('./index')


yargs(hideBin(process.argv))
    .scriptName('wot-testing-dns-sd')
    .command('$0 [file]', 'A simple utility to test a discoverer and a discoveree according to https://www.w3.org/TR/wot-discovery/#introduction-dns-sd', () => { }, async (argv) => {
        const testSuite = new MDNSWoTTest();
        const results = [];

        const file = argv.file ? argv.file : 'dns-sd-test.csv'
        const kind = argv.kind ? argv.kind : '*'

        const resultDiscovererThing = (kind === 'thing' || kind === '*') && testSuite.testDiscovererThing(argv.timeout).then(() => {
            results.push({ id: 'introduction-dns-sd-service-name_discoverer_thing', 'status': 'pass', comment: '' })
        }).catch((e) => {
            results.push({ id: 'introduction-dns-sd-service-name_discoverer_thing', 'status': 'fail', comment: e.message })
        })

        const resultDiscovererDirectory = (kind === 'directory' || kind === '*') && testSuite.testDiscovererDirectory(argv.timeout).then(() => {
            results.push({ id: 'introduction-dns-sd-service-name_discoverer_directory', 'status': 'pass', comment: '' })
        }).catch((e) => {
            results.push({ id: 'introduction-dns-sd-service-name_discoverer_directory', 'status': 'fail', comment: e.message })
        })

        await Promise.all([resultDiscovererThing, resultDiscovererDirectory])

        const resultDiscovereeThing = (kind === 'thing' || kind === '*') && testSuite.testDiscovereeThing(argv.timeout).then(() => {
            results.push({ id: 'introduction-dns-sd-service-name_discoveree_thing', 'status': 'pass', comment: '' })
        }).catch((e) => {
            results.push({ id: 'introduction-dns-sd-service-name_discoveree_thing', 'status': 'fail', comment: e.message })
        })

        const resultDiscovereeDirectory = (kind === 'directory' || kind === '*') && testSuite.testDiscovereeDirectory(argv.timeout).then(() => {
            results.push({ id: 'introduction-dns-sd-service-name_discoveree_directory', 'status': 'pass', comment: '' })
        }).catch((e) => {
            results.push({ id: 'introduction-dns-sd-service-name_discoveree_directory', 'status': 'fail', comment: e.message })
        })

        await Promise.all([resultDiscovereeThing, resultDiscovereeDirectory]);

        await writeResult(results, file, argv.append)

        !argv.quiet && printReports(results)

        testSuite.close()
    })
    .command('discoverer [file]', 'testing a discoverer (whom issues queries)', () => { }, async (argv) => {
        const testSuite = new MDNSWoTTest();
        const results = [];

        const file = argv.file ? argv.file : 'dns-sd-test.csv'
        const kind = argv.kind ? argv.kind : '*'

        const resultDiscovererThing = (kind === 'thing' || kind === '*') && testSuite.testDiscovererThing(argv.timeout).then(() => {
            results.push({ id: 'introduction-dns-sd-service-name_discoverer_thing', 'status': 'pass', comment: '' })
        }).catch((e) => {
            results.push({ id: 'introduction-dns-sd-service-name_discoverer_thing', 'status': 'fail', comment: e.message })
        })

        const resultDiscovererDirectory = (kind === 'directory' || kind === '*') && testSuite.testDiscovererDirectory(argv.timeout).then(() => {
            results.push({ id: 'introduction-dns-sd-service-name_discoverer_directory', 'status': 'pass', comment: '' })
        }).catch((e) => {
            results.push({ id: 'introduction-dns-sd-service-name_discoverer_directory', 'status': 'fail', comment: e.message })
        })

        await Promise.all([resultDiscovererThing, resultDiscovererDirectory])

        await writeResult(results, file, argv.append)

        !argv.quiet && printReports(results)

        testSuite.close()
    })
    .command('discoveree [file]', 'testing an exposed Thing', () => { }, async (argv) => {
        const testSuite = new MDNSWoTTest();
        const results = [];

        const file = argv.file ? argv.file : 'dns-sd-test.csv'
        const kind = argv.kind ? argv.kind : '*'

        const resultDiscovereeThing = (kind === 'thing' || kind === '*') && testSuite.testDiscovereeThing(argv.timeout).then(() => {
            results.push({ id: 'introduction-dns-sd-service-name_discoveree_thing', 'status': 'pass', comment: '' })
        }).catch((e) => {
            results.push({ id: 'introduction-dns-sd-service-name_discoveree_thing', 'status': 'fail', comment: e.message })
        })

        const resultDiscovereeDirectory = (kind === 'directory' || kind === '*') && testSuite.testDiscovereeDirectory(argv.timeout).then(() => {
            results.push({ id: 'introduction-dns-sd-service-name_discoveree_directory', 'status': 'pass', comment: '' })
        }).catch((e) => {
            results.push({ id: 'introduction-dns-sd-service-name_discoveree_directory', 'status': 'fail', comment: e.message })
        })

        await Promise.all([resultDiscovereeThing, resultDiscovereeDirectory]);

        await writeResult(results, file, argv.append)
        
        !argv.quiet && printReports(results)

        testSuite.close()
    })
    .env('WOT_DNS_SD_TESTER')
    .default('append', false)
    .option('append', {
        alias: 'a',
        type: 'boolean',
        description: 'Append the test result to file'
    })
    .default('timeout', 5000)
    .option('timeout', {
        alias: 't',
        type: 'number',
        description: 'Total maximum milliseconds of the test',
    })
    .option('kind', {
        alias: 'k',
        describe: 'Choose to test for Thing Directory or both',
        choices: ['thing', 'directory', '*']
    })
    .option('quiet', {
        alias: 'q',
        describe: 'Do not print results'
    })
    .example('wot_dnssd_tester discover --kind thing test only queries for Thing')

    .argv

async function writeResult(results, file, append) {
    const createCsvWriter = require('csv-writer').createObjectCsvWriter;
    const csvWriter = createCsvWriter({
        path: file,
        header: [
            { id: 'id', title: 'ID' },
            { id: 'status', title: 'Status' },
            { id: 'comment', title: 'Comment' },
        ],
        append: append
    });

    await csvWriter.writeRecords(results)

}

function printReports(results){
    console.log('assertion\t\t\t\t\t\tstatus\tcomment\n')
    for (const report of results) {
        console.log(`${report.id}\t${report.status}\t${report.comment}`)
    }
}

