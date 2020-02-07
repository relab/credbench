const { expectEvent, constants, time, expectRevert, ether } = require('@openzeppelin/test-helpers');

const Faculty = artifacts.require('FacultyMock');
const Course = artifacts.require('CourseMock');

contract('Faculty', accounts => {
    const [dean, adm, teacher, evaluator, student, other] = accounts;
    let faculty, courseStarts, courseEnds = null;
    const semester = web3.utils.keccak256(web3.utils.toHex('spring2020'));
    const digest1 = web3.utils.keccak256(web3.utils.toHex('cert1'));
    const digest2 = web3.utils.keccak256(web3.utils.toHex('cert2'));
    const digest3 = web3.utils.keccak256(web3.utils.toHex('cert3'));
    const courseDigest1 = web3.utils.keccak256(web3.utils.toHex('course1'));
    const courseDigest2 = web3.utils.keccak256(web3.utils.toHex('course2'));
    const diplomaDigest = web3.utils.keccak256(web3.utils.toHex('diploma'));

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
            let courses = [];
            let expectedCerts = [];
            for (i = 0; i < 3; i++) {
                courseStarts = (await time.latest()).add(time.duration.seconds(1));
                courseEnds = courseStarts.add(await time.duration.hours(1));
                await faculty.createCourse(semester, [teacher, evaluator], 2, courseStarts, courseEnds);
                await time.increase(time.duration.seconds(1));
                let course = await Course.at(await faculty.coursesBySemester(semester, i));

                courses.push(course.address);
                await course.addStudent(student, { from: teacher });

                for (d of [digest1, digest2, digest3]) {
                    await course.issueExam(student, d, { from: teacher });
                    await course.issueExam(student, d, { from: evaluator });
                    await course.confirmProof(d, { from: student });
                    await time.increase(time.duration.seconds(1));
                }
            }

            await time.increase(time.duration.hours(1));

            for (cAddress of courses) {
                let course = await Course.at(cAddress);
                (await course.hasEnded()).should.equal(true);

                await course.issueCourseCertificate(student, { from: teacher });
                let cert = (await course.courseCertificate(student)).digest
                expectedCerts.push(cert);
            }

            await time.increase(time.duration.seconds(1));

            await faculty.issueDiploma(student, courses);

            let expected = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', expectedCerts));

            let diploma = await faculty.digestsBySubject(student, 0);

            (diploma).should.equal(expected);
        });
    });
});