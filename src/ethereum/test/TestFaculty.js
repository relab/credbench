const { expectEvent, constants, time, expectRevert, ether } = require('@openzeppelin/test-helpers');

const Faculty = artifacts.require('FacultyMock');
const Course = artifacts.require('CourseMock');

contract('Faculty', accounts => {
    const [dean, adm, teacher, evaluator, student, other] = accounts;
    let faculty, courseStarts, courseEnds = null;
    const semester = web3.utils.keccak256(web3.utils.toHex('spring2020'));

    describe('constructor', () => {
        it('should successfully deploy the contract', async () => {
            faculty = await Faculty.new([dean, adm], 2);
            (await faculty.isOwner(dean)).should.equal(true);
            (await faculty.isOwner(adm)).should.equal(true);
            assert(faculty.quorum(), 2);
        });
    });

    describe('course creation', () => {
        beforeEach(async () => {
            faculty = await Faculty.new([dean, adm], 2);
            faculty.setBalance({ value: ether('1') });
            courseStarts = (await time.latest()).add(time.duration.seconds(1));
            courseEnds = courseStarts.add(await time.duration.hours(1));
            await time.increase(time.duration.seconds(1));
        });

        it('should create a course', async () => {
            const { logs } = await faculty.createCourse(semester, [teacher, evaluator], 2, courseStarts, courseEnds);

            expectEvent.inLogs(logs, 'CourseCreated');
            await time.increase(time.duration.seconds(1));
        });
    });

    describe('issuing diploma', () => {
        beforeEach(async () => {
            faculty = await Faculty.new([dean, adm], 2);
            faculty.setBalance({ value: ether('1') });
        });

        it('should issue a diploma', async () => {
            let courses = new Map();
            for (i = 0; i < 3; i++) {
                // Create the course
                courseStarts = (await time.latest()).add(time.duration.seconds(1 + i));
                courseEnds = courseStarts.add(await time.duration.hours(1));
                await faculty.createCourse(semester, [teacher, evaluator], 2, courseStarts, courseEnds);
                // Start the course
                await time.increase(time.duration.seconds(1 + i));
                // Get the course instance
                let course = await Course.at(await faculty.coursesBySemester(semester, i));
                // Create course digest
                let courseDigest = web3.utils.keccak256(web3.utils.toHex(`course${i}`));
                courses.set(course.address, courseDigest);
                // Add student
                await course.addStudent(student, { from: teacher });
                // Add exam certificates
                for (j = 0; j < 3 + i; j++) {
                    let examDigest = web3.utils.keccak256(web3.utils.toHex(`course${i}-exam${j}`));
                    await course.issueExam(student, examDigest, { from: teacher });
                    await course.issueExam(student, examDigest, { from: evaluator });
                    await course.confirmProof(examDigest, { from: student });
                    await time.increase(time.duration.seconds(1));
                }
            }
            // Finish courses
            await time.increase(time.duration.hours(1));

            var certs = [];
            for (let [cAddress, cDigest] of courses) {
                let course = await Course.at(cAddress);
                (await course.hasEnded()).should.equal(true);

                await course.issueCourseCertificate(student, cDigest, { from: teacher });
                await course.issueCourseCertificate(student, cDigest, { from: evaluator });
                await course.confirmProof(cDigest, { from: student });
                certs.push(await course.digestsBySubject(student));
            }
            // Aggregate courses certs
            var expectedCerts = certs.map(c => web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', c)));

            // Add diploma digest
            let diplomaDigest = web3.utils.keccak256(web3.utils.toHex('diploma'));
            // Add new diploma to certs
            expectedCerts.push(diplomaDigest);
            // Create new root
            let expectedRoot = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', expectedCerts));

            await faculty.issueDiploma(student, diplomaDigest, expectedRoot, Array.from(courses.keys()), { from: adm });

            let diploma = (await faculty.digestsBySubject(student))[0];
            (diploma).should.equal(diplomaDigest);
        });
    });
});